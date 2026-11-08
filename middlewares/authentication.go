package middlewares

import (
	"errors"
	"log/slog"
	"slices"

	"gateway/db"
	"gateway/internal/crypto"
	"gateway/schema"

	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/gofiber/fiber/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const APIKeyHeader = "X-API-KEY"

type Authenticator struct {
	store   *db.Store
	auth    auth.AuthServiceClient
	cookie  SessionCookie
	origins []string
}

func NewAuthenticator(store *db.Store, authClient auth.AuthServiceClient, cookie SessionCookie, allowedOrigins []string) *Authenticator {
	return &Authenticator{store: store, auth: authClient, cookie: cookie, origins: slices.Clone(allowedOrigins)}
}

func (a *Authenticator) Required() fiber.Handler { return a.handler(false) }

func (a *Authenticator) AllowPendingPassword() fiber.Handler { return a.handler(true) }

func (a *Authenticator) handler(allowPending bool) fiber.Handler {
	return func(c fiber.Ctx) error {
		if key := c.Get(APIKeyHeader); key != "" {
			return a.apiKey(c, key)
		}
		if value := a.cookie.Read(c); value != "" {
			return a.session(c, value, allowPending)
		}
		return fiber.ErrUnauthorized
	}
}

func (a *Authenticator) session(c fiber.Ctx, value string, allowPending bool) error {
	userID, sid, ok := db.ParseCookie(value)
	if !ok {
		a.cookie.Clear(c)
		return fiber.NewError(fiber.StatusUnauthorized, "Session expired")
	}

	session, err := a.store.TouchSession(c.Context(), userID, sid)
	if errors.Is(err, db.ErrInvalidSession) {
		a.cookie.Clear(c)
		return fiber.NewError(fiber.StatusUnauthorized, "Session expired")
	}
	if err != nil {
		slog.LogAttrs(c.Context(), slog.LevelError, "reading session", slog.Any("error", err))
		return fiber.NewError(fiber.StatusServiceUnavailable, "Session store unavailable")
	}

	identity := session.Identity
	if identity == nil {
		identity, err = a.refreshIdentity(c, userID)
		if err != nil {
			return err
		}
	}

	if !isSafeMethod(c.Method()) && !slices.Contains(a.origins, c.Get(fiber.HeaderOrigin)) {
		slog.LogAttrs(c.Context(), slog.LevelWarn, "rejected session request from disallowed origin",
			slog.String("origin", c.Get(fiber.HeaderOrigin)), slog.String("path", c.Path()))
		return fiber.NewError(fiber.StatusForbidden, "Origin not allowed")
	}

	a.cookie.Set(c, value)
	c.Locals("auth", &schema.AuthInfo{Identity: *identity, SessionID: sid})

	if identity.MustChangePassword && !allowPending {
		return fiber.NewError(fiber.StatusForbidden, "Password change required")
	}
	return c.Next()
}

func (a *Authenticator) refreshIdentity(c fiber.Ctx, userID int) (*schema.Identity, error) {
	res, err := a.auth.GetIdentity(c.Context(), &auth.UserRequest{UserId: int32(userID)})
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound, codes.Unauthenticated:
			if err := a.store.DeleteAllSessions(c.Context(), userID); err != nil {
				slog.LogAttrs(c.Context(), slog.LevelError, "revoking sessions of removed user", slog.Any("error", err))
			}
			a.cookie.Clear(c)
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Session expired")
		default:
			slog.LogAttrs(c.Context(), slog.LevelError, "loading identity", slog.Any("error", err))
			return nil, fiber.NewError(fiber.StatusServiceUnavailable, "Auth service unavailable")
		}
	}

	identity := schema.IdentityFrom(res)
	if err := a.store.CacheIdentity(c.Context(), identity); err != nil {
		slog.LogAttrs(c.Context(), slog.LevelWarn, "caching identity", slog.Any("error", err))
	}
	return &identity, nil
}

func (a *Authenticator) apiKey(c fiber.Ctx, key string) error {
	hash := crypto.HashSHA256(key)

	identity, err := a.store.GetAPIKey(c.Context(), hash)
	switch {
	case errors.Is(err, db.ErrInvalidSession):
		identity, err = a.validateKey(c, key, hash)
		if err != nil {
			return err
		}
	case err != nil:
		slog.LogAttrs(c.Context(), slog.LevelError, "reading api key cache", slog.Any("error", err))
		return fiber.NewError(fiber.StatusServiceUnavailable, "Session store unavailable")
	}

	if origin := c.Get(fiber.HeaderOrigin); origin != "" && !slices.Contains(identity.Origins, origin) {
		return fiber.NewError(fiber.StatusForbidden, "Origin not allowed for this API key")
	}

	c.Locals("auth", &schema.AuthInfo{Identity: identity.Identity, APIKey: true})
	return c.Next()
}

func (a *Authenticator) validateKey(c fiber.Ctx, key, hash string) (schema.APIKeyIdentity, error) {
	res, err := a.auth.ValidateKey(c.Context(), &auth.ValidateKeyRequest{Key: key})
	if err != nil {
		if code := status.Code(err); code == codes.NotFound || code == codes.InvalidArgument {
			return schema.APIKeyIdentity{}, fiber.NewError(fiber.StatusUnauthorized, "Invalid API key")
		}
		slog.LogAttrs(c.Context(), slog.LevelError, "validating api key", slog.Any("error", err))
		return schema.APIKeyIdentity{}, fiber.NewError(fiber.StatusServiceUnavailable, "Auth service unavailable")
	}

	origins := res.GetOrigins()
	if origins == nil {
		origins = []string{}
	}
	identity := schema.APIKeyIdentity{Identity: schema.IdentityFrom(res.GetIdentity()), Origins: origins}
	if err := a.store.CacheAPIKey(c.Context(), hash, identity); err != nil {
		slog.LogAttrs(c.Context(), slog.LevelWarn, "caching api key", slog.Any("error", err))
	}
	return identity, nil
}

func isSafeMethod(method string) bool {
	return method == fiber.MethodGet || method == fiber.MethodHead || method == fiber.MethodOptions
}
