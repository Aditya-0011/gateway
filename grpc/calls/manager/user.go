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

func PortfolioUserCalls(redis *db.RedisParams, client manager.UserServiceClient, validator protovalidate.Validator) *PortfolioUserCallsParams {
	return &PortfolioUserCallsParams{
		redis:     redis,
		client:    client,
		validator: validator,
	}
}

func (p *PortfolioUserCallsParams) GetUserDetails(c fiber.Ctx) error {
	userId := c.Locals("userId").(int)

	req := &manager.SimpleRequest{
		UserId: int32(userId),
	}

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.GetUserDetailsResponse, error) {
		return p.client.GetUserDetails(ctx, req)
	})
}

func (p *PortfolioUserCallsParams) EditUserDetails(c fiber.Ctx) error {
	userId := c.Locals("userId").(int)

	req := c.Locals("req").(*manager.EditUserDetailsRequest)
	req.UserId = int32(userId)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*manager.SimpleResponse, error) {
		return p.client.EditUserDetails(ctx, req)
	})
}
