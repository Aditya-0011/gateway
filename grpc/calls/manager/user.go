package calls

import (
	"context"
	rpc "gateway/grpc/calls"

	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/gofiber/fiber/v3"
)

type PortfolioUserCallsParams struct {
	client manager.UserServiceClient
}

func PortfolioUserCalls(client manager.UserServiceClient) *PortfolioUserCallsParams {
	return &PortfolioUserCallsParams{
		client: client,
	}
}

func (p *PortfolioUserCallsParams) GetUserDetails(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.SimpleRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.GetUserDetailsResponse, error) {
		return p.client.GetUserDetails(ctx, req)
	})
}

func (p *PortfolioUserCallsParams) EditUserDetails(c fiber.Ctx) error {
	req := c.Locals("req").(*manager.EditUserDetailsRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.SimpleResponse, error) {
		return p.client.EditUserDetails(ctx, req)
	})
}
