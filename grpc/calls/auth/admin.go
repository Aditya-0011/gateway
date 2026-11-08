package calls

import (
	"context"

	rpc "gateway/grpc/calls"
	"gateway/internal/wire"

	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/gofiber/fiber/v3"
)

func (h *Handlers) ListUsers(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.UserRequest)
	return rpc.CallWithJSON(c, func(ctx context.Context) (*auth.UsersResponse, error) {
		return h.client.ListUsers(ctx, req)
	})
}

func (h *Handlers) CreateUser(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.CreateUserRequest)
	return rpc.CallWithJSON(c, func(ctx context.Context) (*auth.CreateUserResponse, error) {
		return h.client.CreateUser(ctx, req)
	})
}

func (h *Handlers) ResetUserPassword(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.TargetUserRequest)

	res, err := rpc.Call(c, func(ctx context.Context) (*auth.TemporaryPasswordResponse, error) {
		return h.client.ResetUserPassword(ctx, req)
	})
	if err != nil {
		return err
	}
	h.warn(c, "revoking sessions after admin reset", h.store.DeleteAllSessions(c.Context(), int(req.GetTargetUserId())))
	return wire.Write(c, res)
}

func (h *Handlers) SetUserStatus(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.SetUserStatusRequest)

	res, err := rpc.Call(c, func(ctx context.Context) (*auth.SimpleResponse, error) {
		return h.client.SetUserStatus(ctx, req)
	})
	if err != nil {
		return err
	}
	target := int(req.GetTargetUserId())
	if req.GetStatus() == auth.UserStatus_USER_STATUS_DISABLED {
		h.warn(c, "revoking sessions of disabled user", h.store.DeleteAllSessions(c.Context(), target))
	} else {
		h.warn(c, "invalidating identity", h.store.InvalidateIdentity(c.Context(), target))
	}
	return wire.Write(c, res)
}

func (h *Handlers) SetUserProjects(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.SetUserProjectsRequest)

	res, err := rpc.Call(c, func(ctx context.Context) (*auth.SimpleResponse, error) {
		return h.client.SetUserProjects(ctx, req)
	})
	if err != nil {
		return err
	}
	h.warn(c, "invalidating identity", h.store.InvalidateIdentity(c.Context(), int(req.GetTargetUserId())))
	return wire.Write(c, res)
}

func (h *Handlers) DeleteUser(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.TargetUserRequest)

	res, err := rpc.Call(c, func(ctx context.Context) (*auth.SimpleResponse, error) {
		return h.client.DeleteUser(ctx, req)
	})
	if err != nil {
		return err
	}
	h.warn(c, "revoking sessions of deleted user", h.store.DeleteAllSessions(c.Context(), int(req.GetTargetUserId())))
	return wire.Write(c, res)
}

func (h *Handlers) ListProjects(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.UserRequest)
	return rpc.CallWithJSON(c, func(ctx context.Context) (*auth.ProjectsResponse, error) {
		return h.client.ListProjects(ctx, req)
	})
}

func (h *Handlers) CreateProject(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.SaveProjectRequest)
	req.Id = 0
	return h.saveProject(c, req)
}

func (h *Handlers) UpdateProject(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.SaveProjectRequest)
	if req.GetId() <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "id: value must be greater than 0")
	}
	return h.saveProject(c, req)
}

func (h *Handlers) saveProject(c fiber.Ctx, req *auth.SaveProjectRequest) error {
	res, err := rpc.Call(c, func(ctx context.Context) (*auth.SimpleResponse, error) {
		return h.client.SaveProject(ctx, req)
	})
	if err != nil {
		return err
	}
	h.warn(c, "invalidating identities", h.store.InvalidateAllIdentities(c.Context()))
	return wire.Write(c, res)
}

func (h *Handlers) DeleteProject(c fiber.Ctx) error {
	req := c.Locals("req").(*auth.DeleteProjectRequest)

	res, err := rpc.Call(c, func(ctx context.Context) (*auth.SimpleResponse, error) {
		return h.client.DeleteProject(ctx, req)
	})
	if err != nil {
		return err
	}
	h.warn(c, "invalidating identities", h.store.InvalidateAllIdentities(c.Context()))
	return wire.Write(c, res)
}
