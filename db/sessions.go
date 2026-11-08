package db

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	"gateway/internal/crypto"
	"gateway/schema"

	"github.com/bytedance/sonic"
	driver "github.com/redis/go-redis/v9"
)

const (
	SessionIdleTimeout = 15 * time.Minute
	SessionMaxLifetime = 12 * time.Hour
	authzTTL           = 10 * time.Minute
	authzUsersKey      = "authz:users"
	userAgentMaxLen    = 200
)

var (
	createSessionScript = driver.NewScript(`
		redis.call("HSET", KEYS[1], "created", ARGV[1], "last", ARGV[1], "ip", ARGV[2], "ua", ARGV[3])
		redis.call("EXPIRE", KEYS[1], ARGV[4])
		redis.call("SADD", KEYS[2], ARGV[6])
		redis.call("EXPIRE", KEYS[2], ARGV[5])
		return 1
	`)

	touchSessionScript = driver.NewScript(`
		local created = redis.call("HGET", KEYS[1], "created")
		if not created then
			return false
		end
		if tonumber(ARGV[1]) - tonumber(created) > tonumber(ARGV[3]) then
			redis.call("DEL", KEYS[1])
			redis.call("SREM", KEYS[2], ARGV[4])
			return false
		end
		redis.call("HSET", KEYS[1], "last", ARGV[1])
		redis.call("EXPIRE", KEYS[1], ARGV[2])
		local authz = redis.call("GET", KEYS[3])
		return {created, authz or ""}
	`)

	listSessionsScript = driver.NewScript(`
		local result = {}
		for _, sid in ipairs(redis.call("SMEMBERS", KEYS[1])) do
			local fields = redis.call("HMGET", ARGV[1] .. sid, "created", "last", "ip", "ua")
			if fields[1] then
				for _, value in ipairs({sid, fields[1], fields[2] or "", fields[3] or "", fields[4] or ""}) do
					table.insert(result, value)
				end
			else
				redis.call("SREM", KEYS[1], sid)
			end
		end
		return result
	`)

	deleteSessionScript = driver.NewScript(`
		redis.call("DEL", KEYS[1])
		redis.call("SREM", KEYS[2], ARGV[1])
		return 1
	`)

	deleteSessionsScript = driver.NewScript(`
		local removed = 0
		for _, sid in ipairs(redis.call("SMEMBERS", KEYS[1])) do
			if sid ~= ARGV[2] then
				redis.call("DEL", ARGV[1] .. sid)
				redis.call("SREM", KEYS[1], sid)
				removed = removed + 1
			end
		end
		if ARGV[2] == "" then
			redis.call("DEL", KEYS[1], KEYS[2])
		end
		return removed
	`)
)

type Session struct {
	UserID    int
	ID        string
	CreatedAt time.Time
	Identity  *schema.Identity
}

type SessionMeta struct {
	IP        string
	UserAgent string
}

func userTag(userID int) string { return "session:{u" + strconv.Itoa(userID) + "}:" }

func sessionKey(userID int, sid string) string { return userTag(userID) + sid }

func indexKey(userID int) string { return userTag(userID) + "index" }

func authzKey(userID int) string { return userTag(userID) + "authz" }

func sessionID(token string) string { return crypto.HashSHA256(token) }

func (s *Store) CreateSession(ctx context.Context, userID int, meta SessionMeta) (cookie string, sid string, err error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	token, err := crypto.NewSessionKey()
	if err != nil {
		return "", "", err
	}
	sid = sessionID(token)

	ua := meta.UserAgent
	if len(ua) > userAgentMaxLen {
		ua = ua[:userAgentMaxLen]
	}

	err = createSessionScript.Run(ctx, s.client,
		[]string{sessionKey(userID, sid), indexKey(userID)},
		time.Now().Unix(), meta.IP, ua,
		int(SessionIdleTimeout.Seconds()), int(SessionMaxLifetime.Seconds()), sid,
	).Err()
	if err != nil {
		return "", "", err
	}
	return strconv.Itoa(userID) + "." + token, sid, nil
}

func ParseCookie(value string) (userID int, sid string, ok bool) {
	for i := 0; i < len(value); i++ {
		if value[i] == '.' {
			id, err := strconv.Atoi(value[:i])
			token := value[i+1:]
			if err != nil || id <= 0 || len(token) < 32 {
				return 0, "", false
			}
			return id, sessionID(token), true
		}
	}
	return 0, "", false
}

func (s *Store) TouchSession(ctx context.Context, userID int, sid string) (Session, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	res, err := touchSessionScript.Run(ctx, s.client,
		[]string{sessionKey(userID, sid), indexKey(userID), authzKey(userID)},
		time.Now().Unix(), int(SessionIdleTimeout.Seconds()), int(SessionMaxLifetime.Seconds()), sid,
	).StringSlice()
	if errors.Is(err, driver.Nil) {
		return Session{}, ErrInvalidSession
	}
	if err != nil {
		return Session{}, err
	}

	created, _ := strconv.ParseInt(res[0], 10, 64)
	session := Session{UserID: userID, ID: sid, CreatedAt: time.Unix(created, 0)}
	if res[1] != "" {
		var identity schema.Identity
		if err := sonic.UnmarshalString(res[1], &identity); err == nil && identity.UserID == userID {
			session.Identity = &identity
		}
	}
	return session, nil
}

func (s *Store) CacheIdentity(ctx context.Context, identity schema.Identity) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	data, err := sonic.Marshal(identity)
	if err != nil {
		return err
	}
	pipe := s.client.Pipeline()
	pipe.Set(ctx, authzKey(identity.UserID), data, authzTTL)
	pipe.SAdd(ctx, authzUsersKey, identity.UserID)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) InvalidateIdentity(ctx context.Context, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	keys := []string{authzKey(userID), keyOriginsKey}
	if hash, err := s.client.Get(ctx, apiKeyPointer(userID)).Result(); err == nil {
		keys = append(keys, apiKeyPrefix+hash, apiKeyPointer(userID))
	} else if !errors.Is(err, driver.Nil) {
		return err
	}
	return s.client.Del(ctx, keys...).Err()
}

func (s *Store) InvalidateAllIdentities(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	ids, err := s.client.SMembers(ctx, authzUsersKey).Result()
	if err != nil {
		return err
	}
	userIDs := make([]int, 0, len(ids))
	for _, raw := range ids {
		if id, err := strconv.Atoi(raw); err == nil {
			userIDs = append(userIDs, id)
		}
	}

	lookups := make([]*driver.StringCmd, len(userIDs))
	if len(userIDs) > 0 {
		read := s.client.Pipeline()
		for i, id := range userIDs {
			lookups[i] = read.Get(ctx, apiKeyPointer(id))
		}
		if _, err := read.Exec(ctx); err != nil && !errors.Is(err, driver.Nil) {
			return err
		}
	}

	pipe := s.client.Pipeline()
	pipe.Del(ctx, keyOriginsKey)
	for _, raw := range ids {
		pipe.SRem(ctx, authzUsersKey, raw)
	}
	for i, id := range userIDs {
		pipe.Del(ctx, authzKey(id), apiKeyPointer(id))
		if hash, err := lookups[i].Result(); err == nil {
			pipe.Del(ctx, apiKeyPrefix+hash)
		}
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) ListSessions(ctx context.Context, userID int, currentSID string) ([]schema.SessionInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	fields, err := listSessionsScript.Run(ctx, s.client, []string{indexKey(userID)}, userTag(userID)).StringSlice()
	if err != nil {
		return nil, err
	}

	const fieldsPerSession = 5
	sessions := make([]schema.SessionInfo, 0, len(fields)/fieldsPerSession)
	for f := fields; len(f) >= fieldsPerSession; f = f[fieldsPerSession:] {
		sessions = append(sessions, schema.SessionInfo{
			ID:         f[0],
			Current:    f[0] == currentSID,
			CreatedAt:  unixToRFC3339(f[1]),
			LastSeenAt: unixToRFC3339(f[2]),
			IP:         f[3],
			UserAgent:  f[4],
		})
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].LastSeenAt > sessions[j].LastSeenAt })
	return sessions, nil
}

func (s *Store) DeleteSession(ctx context.Context, userID int, sid string) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	return deleteSessionScript.Run(ctx, s.client, []string{sessionKey(userID, sid), indexKey(userID)}, sid).Err()
}

func (s *Store) DeleteOtherSessions(ctx context.Context, userID int, keepSID string) error {
	return s.deleteSessions(ctx, userID, keepSID)
}

func (s *Store) DeleteAllSessions(ctx context.Context, userID int) error {
	return s.deleteSessions(ctx, userID, "")
}

func (s *Store) deleteSessions(ctx context.Context, userID int, keepSID string) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	return deleteSessionsScript.Run(ctx, s.client,
		[]string{indexKey(userID), authzKey(userID)}, userTag(userID), keepSID,
	).Err()
}

func unixToRFC3339(v string) string {
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return ""
	}
	return time.Unix(n, 0).UTC().Format(time.RFC3339)
}
