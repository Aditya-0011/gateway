package db

import (
	"context"
	"errors"
	"math/rand/v2"
	"os"
	"testing"

	"gateway/schema"
)

func TestRevokeUserDropsSessionsAndAPIKey(t *testing.T) {
	url := os.Getenv("GATEWAY_TEST_REDIS_URL")
	if url == "" {
		t.Skip("GATEWAY_TEST_REDIS_URL is not set")
	}
	ctx := context.Background()
	store, err := Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	userID := 1_000_000 + rand.IntN(1_000_000_000)
	identity := schema.Identity{UserID: userID, Projects: []string{}}
	hash := sessionID("key-" + t.Name())

	if err := store.CacheAPIKey(ctx, hash, schema.APIKeyIdentity{Identity: identity, Origins: []string{}}); err != nil {
		t.Fatal(err)
	}
	if err := store.CacheIdentity(ctx, identity); err != nil {
		t.Fatal(err)
	}
	_, sid, err := store.CreateSession(ctx, userID, SessionMeta{})
	if err != nil {
		t.Fatal(err)
	}

	if err := store.RevokeUser(ctx, userID); err != nil {
		t.Fatal(err)
	}

	if _, err := store.GetAPIKey(ctx, hash); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("GetAPIKey after revoke: err = %v, want ErrInvalidSession", err)
	}
	if _, err := store.TouchSession(ctx, userID, sid); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("TouchSession after revoke: err = %v, want ErrInvalidSession", err)
	}
	if n, err := store.client.Exists(ctx, authzKey(userID), apiKeyPointer(userID)).Result(); err != nil || n != 0 {
		t.Fatalf("authz or api key pointer left behind: n = %d, err = %v", n, err)
	}
}
