package calls

import (
	"context"
	"log/slog"
	"regexp"
	"strings"

	"gateway/db"
	rpc "gateway/grpc/calls"
	"gateway/internal/wire"
	"gateway/middlewares"
	"gateway/schema"

	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
)

var sessionIDPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type Handlers struct {
	store  *db.Store
	client auth.AuthServiceClient
	cookie middlewares.SessionCookie
}

func New(store *db.Store, client auth.AuthServiceClient, cookie middlewares.SessionCookie) *Handlers {
	return &Handlers{store: store, client: client, cookie: cookie}
}

func info(c fiber.Ctx) *schema.AuthInfo {
	return c.Locals("auth").(*schema.AuthInfo)
}

func message(c fiber.Ctx, text string) error {
	return c.JSON(schema.MessageResponse{Message: text})
}

func (h *Handlers) warn(c fiber.Ctx, action string, err error) {
	if err != nil {
		slog.LogAttrs(c.Context(), slog.LevelError, action, slog.Any("error", err))
	}
}

func (h *Handlers) Login(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.LoginRequest)

	res, err := rpc.Call(c, func(ctx context.Context) (*auth.Identity, error) {
		return h.client.Login(ctx, req)
	})
	if err != nil {
		return err
	}

	identity := schema.IdentityFrom(res)
	cookie, sid, err := h.store.CreateSession(c.Context(), identity.UserID, db.SessionMeta{
		IP:        c.IP(),
		UserAgent: strings.Clone(c.Get(fiber.HeaderUserAgent)),
	})
	if err != nil {
		h.warn(c, "creating session", err)
		return fiber.NewError(fiber.StatusServiceUnavailable, "Session store unavailable")
	}
	h.warn(c, "caching identity", h.store.CacheIdentity(c.Context(), identity))

	h.cookie.Set(c, cookie)
	return c.JSON(schema.SessionResponse{Identity: identity, SessionID: sid})
}

func (h *Handlers) Session(c fiber.Ctx) error {
	i := info(c)
	return c.JSON(schema.SessionResponse{Identity: i.Identity, SessionID: i.SessionID})
}

func (h *Handlers) Logout(c fiber.Ctx) error {
	i := info(c)
	if err := h.store.DeleteSession(c.Context(), i.UserID, i.SessionID); err != nil {
		h.warn(c, "deleting session", err)
		return fiber.NewError(fiber.StatusServiceUnavailable, "Logout failed, try again")
	}
	h.cookie.Clear(c)
	return message(c, "Successfully logged out")
}

func (h *Handlers) FirstLogin(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.CompleteFirstLoginRequest)
	i := info(c)
	if !i.MustChangePassword {
		return fiber.NewError(fiber.StatusPreconditionFailed, "Password is already set")
	}

	res, err := rpc.Call(c, func(ctx context.Context) (*auth.RecoveryCodesResponse, error) {
		return h.client.CompleteFirstLogin(ctx, req)
	})
	if err != nil {
		return err
	}
	h.warn(c, "invalidating identity", h.store.InvalidateIdentity(c.Context(), i.UserID))
	h.warn(c, "revoking other sessions", h.store.DeleteOtherSessions(c.Context(), i.UserID, i.SessionID))
	return wire.Write(c, res)
}

func (h *Handlers) ChangePassword(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.ChangePasswordRequest)
	i := info(c)

	res, err := rpc.Call(c, func(ctx context.Context) (*auth.SimpleResponse, error) {
		return h.client.ChangePassword(ctx, req)
	})
	if err != nil {
		return err
	}
	h.warn(c, "revoking other sessions", h.store.DeleteOtherSessions(c.Context(), i.UserID, i.SessionID))
	return wire.Write(c, res)
}

func (h *Handlers) ResetPassword(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.ResetPasswordRequest)

	res, err := rpc.Call(c, func(ctx context.Context) (*auth.Identity, error) {
		return h.client.ResetPassword(ctx, req)
	})
	if err != nil {
		return err
	}

	if err := h.store.DeleteAllSessions(c.Context(), int(res.GetUserId())); err != nil {
		h.warn(c, "revoking sessions after password reset", err)
		return fiber.NewError(fiber.StatusServiceUnavailable, "Password changed, but existing sessions could not be revoked")
	}
	h.cookie.Clear(c)
	return message(c, "Password reset. Log in with your new password.")
}

func (h *Handlers) RecoveryCodeStatus(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.UserRequest)
	return rpc.CallWithJSON(c, func(ctx context.Context) (*auth.RecoveryCodeStatusResponse, error) {
		return h.client.GetRecoveryCodeStatus(ctx, req)
	})
}

func (h *Handlers) GenerateRecoveryCodes(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.PasswordConfirmRequest)
	return rpc.CallWithJSON(c, func(ctx context.Context) (*auth.RecoveryCodesResponse, error) {
		return h.client.GenerateRecoveryCodes(ctx, req)
	})
}

func (h *Handlers) DeleteRecoveryCodes(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.PasswordConfirmRequest)
	return rpc.CallWithJSON(c, func(ctx context.Context) (*auth.SimpleResponse, error) {
		return h.client.DeleteRecoveryCodes(ctx, req)
	})
}

func (h *Handlers) Projects(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.UserRequest)
	return rpc.CallWithJSON(c, func(ctx context.Context) (*auth.ProjectsResponse, error) {
		return h.client.ListMyProjects(ctx, req)
	})
}

func (h *Handlers) ListSessions(c fiber.Ctx) error {
	i := info(c)
	sessions, err := h.store.ListSessions(c.Context(), i.UserID, i.SessionID)
	if err != nil {
		h.warn(c, "listing sessions", err)
		return fiber.NewError(fiber.StatusServiceUnavailable, "Session store unavailable")
	}
	return c.JSON(schema.SessionsResponse{Sessions: sessions})
}

func (h *Handlers) RevokeSession(c fiber.Ctx) error {
	if !strings.HasPrefix(c.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
		return fiber.NewError(fiber.StatusUnsupportedMediaType, "Content-Type must be application/json")
	}
	var req schema.RevokeSessionRequest
	if err := sonic.Unmarshal(c.Body(), &req); err != nil || !sessionIDPattern.MatchString(req.SessionID) {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid session id")
	}

	i := info(c)
	if err := h.store.DeleteSession(c.Context(), i.UserID, req.SessionID); err != nil {
		h.warn(c, "revoking session", err)
		return fiber.NewError(fiber.StatusServiceUnavailable, "Session store unavailable")
	}
	if req.SessionID == i.SessionID {
		h.cookie.Clear(c)
	}
	return message(c, "Session signed out")
}

func (h *Handlers) RevokeOtherSessions(c fiber.Ctx) error {
	i := info(c)
	if err := h.store.DeleteOtherSessions(c.Context(), i.UserID, i.SessionID); err != nil {
		h.warn(c, "revoking other sessions", err)
		return fiber.NewError(fiber.StatusServiceUnavailable, "Session store unavailable")
	}
	return message(c, "Signed out of all other devices")
}

func (h *Handlers) GetKey(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.UserRequest)
	return rpc.CallWithJSON(c, func(ctx context.Context) (*auth.GetKeyResponse, error) {
		return h.client.GetKey(ctx, req)
	})
}

func (h *Handlers) RotateKey(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.UserRequest)

	res, err := rpc.Call(c, func(ctx context.Context) (*auth.RotateKeyResponse, error) {
		return h.client.RotateKey(ctx, req)
	})
	if err != nil {
		return err
	}
	if prev := res.GetPreviousKeyHash(); prev != "" {
		h.warn(c, "evicting rotated api key", h.store.EvictAPIKey(c.Context(), prev))
	}
	return wire.Write(c, &auth.RotateKeyResponse{Key: res.GetKey(), UpdatedAt: res.GetUpdatedAt()})
}

func (h *Handlers) RevokeKey(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.UserRequest)

	res, err := rpc.Call(c, func(ctx context.Context) (*auth.RevokeKeyResponse, error) {
		return h.client.RevokeKey(ctx, req)
	})
	if err != nil {
		return err
	}
	if prev := res.GetPreviousKeyHash(); prev != "" {
		h.warn(c, "evicting revoked api key", h.store.EvictAPIKey(c.Context(), prev))
	}
	return message(c, "API key revoked")
}

func (h *Handlers) SetKeyOrigins(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.SetKeyOriginsRequest)

	res, err := rpc.Call(c, func(ctx context.Context) (*auth.KeyOriginsResponse, error) {
		return h.client.SetKeyOrigins(ctx, req)
	})
	if err != nil {
		return err
	}
	h.warn(c, "invalidating api key cache", h.store.InvalidateIdentity(c.Context(), int(req.GetUserId())))
	return wire.Write(c, res)
}
