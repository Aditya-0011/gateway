package calls

import (
	"context"
	rpc "gateway/grpc/calls"

	"github.com/Aditya-0011/common/contracts/go/wallet"
	"github.com/gofiber/fiber/v3"
)

type WalletUserCallsParams struct {
	client wallet.UserServiceClient
}

func WalletUserCalls(client wallet.UserServiceClient) *WalletUserCallsParams {
	return &WalletUserCallsParams{
		client: client,
	}
}

func (w *WalletUserCallsParams) GetUserSummary(c fiber.Ctx) error {
	req := c.Locals("req").(*wallet.SimpleRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*wallet.GetUserSummaryResponse, error) {
		return w.client.GetUserSummary(ctx, req)
	})
}
