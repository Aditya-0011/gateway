package calls

import (
	"context"
	rpc "gateway/grpc/calls"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/wallet"
	"github.com/gofiber/fiber/v3"
)

type WalletUserCallsParams struct {
	client    wallet.UserServiceClient
	validator protovalidate.Validator
}

func WalletUserCalls(client wallet.UserServiceClient, validator protovalidate.Validator) *WalletUserCallsParams {
	return &WalletUserCallsParams{
		client:    client,
		validator: validator,
	}
}

func (w *WalletUserCallsParams) GetUserSummary(c fiber.Ctx) error {
	req := c.Locals("req").(*wallet.SimpleRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*wallet.GetUserSummaryResponse, error) {
		return w.client.GetUserSummary(ctx, req)
	})
}
