package calls

import (
	"context"
	rpc "gateway/grpc/calls"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/gofiber/fiber/v3"
)

type PortfolioCallsParams struct {
	client    manager.PortfolioServiceClient
	validator protovalidate.Validator
}

func PortfolioCalls(client manager.PortfolioServiceClient, validator protovalidate.Validator) *PortfolioCallsParams {
	return &PortfolioCallsParams{
		client:    client,
		validator: validator,
	}
}

func (p *PortfolioCallsParams) GetMessages(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.SimpleRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.GetMessagesResponse, error) {
		return p.client.GetMessages(ctx, req)
	})
}

func (p *PortfolioCallsParams) AddMessage(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.AddMessageRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.SimpleResponse, error) {
		return p.client.AddMessage(ctx, req)
	})
}

func (p *PortfolioCallsParams) DeleteMessage(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.DeleteMessageRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.SimpleResponse, error) {
		return p.client.DeleteMessage(ctx, req)
	})
}

func (p *PortfolioCallsParams) GetTechnologies(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.SimpleRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.GetTechnologiesResponse, error) {
		return p.client.GetTechnologies(ctx, req)
	})
}

func (p *PortfolioCallsParams) CreateTechnology(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.TechnologyCreateRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.SimpleResponse, error) {
		return p.client.CreateTechnology(ctx, req)
	})
}

func (p *PortfolioCallsParams) UpdateTechnology(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.TechnologyUpdateRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.SimpleResponse, error) {
		return p.client.UpdateTechnology(ctx, req)
	})
}

func (p *PortfolioCallsParams) DeleteTechnology(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.DeleteRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.SimpleResponse, error) {
		return p.client.DeleteTechnology(ctx, req)
	})
}

func (p *PortfolioCallsParams) GetProjects(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.SimpleRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.GetProjectsResponse, error) {
		return p.client.GetProjects(ctx, req)
	})
}

func (p *PortfolioCallsParams) CreateProject(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.ProjectCreateRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.SimpleResponse, error) {
		return p.client.CreateProject(ctx, req)
	})
}

func (p *PortfolioCallsParams) UpdateProject(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.ProjectUpdateRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.SimpleResponse, error) {
		return p.client.UpdateProject(ctx, req)
	})
}

func (p *PortfolioCallsParams) DeleteProject(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.DeleteRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.SimpleResponse, error) {
		return p.client.DeleteProject(ctx, req)
	})
}

func (p *PortfolioCallsParams) GetExperiences(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.SimpleRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.GetExperiencesResponse, error) {
		return p.client.GetExperiences(ctx, req)
	})
}

func (p *PortfolioCallsParams) CreateExperience(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.ExperienceCreateRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.SimpleResponse, error) {
		return p.client.CreateExperience(ctx, req)
	})
}

func (p *PortfolioCallsParams) UpdateExperience(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.ExperienceUpdateRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.SimpleResponse, error) {
		return p.client.UpdateExperience(ctx, req)
	})
}

func (p *PortfolioCallsParams) DeleteExperience(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.DeleteRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.SimpleResponse, error) {
		return p.client.DeleteExperience(ctx, req)
	})
}
