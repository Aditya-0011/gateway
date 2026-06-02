package calls

import (
	"context"
	"gateway/db"
	rpc "gateway/grpc/calls"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/gofiber/fiber/v3"
)

type PortfolioUserCallsParams struct {
	redis     *db.RedisParams
	client    manager.UserServiceClient
	validator protovalidate.Validator
}

func PortfolioUserCalls(client manager.UserServiceClient, validator protovalidate.Validator) *PortfolioUserCallsParams {
	return &PortfolioUserCallsParams{
		client:    client,
		validator: validator,
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
