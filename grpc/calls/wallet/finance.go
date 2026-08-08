package calls

import (
	"context"
	rpc "gateway/grpc/calls"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/wallet"
	"github.com/gofiber/fiber/v3"
)

type FinanceCallsParams struct {
	client    wallet.FinanceServiceClient
	validator protovalidate.Validator
}

func FinanceCalls(client wallet.FinanceServiceClient, validator protovalidate.Validator) *FinanceCallsParams {
	return &FinanceCallsParams{
		client:    client,
		validator: validator,
	}
}

func (w *FinanceCallsParams) GetCategories(c fiber.Ctx) error {
	req := c.Locals("req").(*wallet.SimpleRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*wallet.GetCategoriesResponse, error) {
		return w.client.GetCategories(ctx, req)
	})
}

func (w *FinanceCallsParams) CreateCategory(c fiber.Ctx) error {
	req := c.Locals("req").(*wallet.CreateCategoryRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*wallet.SimpleResponse, error) {
		return w.client.CreateCategory(ctx, req)
	})
}

func (w *FinanceCallsParams) UpdateCategory(c fiber.Ctx) error {
	req := c.Locals("req").(*wallet.UpdateCategoryRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*wallet.SimpleResponse, error) {
		return w.client.UpdateCategory(ctx, req)
	})
}

func (w *FinanceCallsParams) DeleteCategory(c fiber.Ctx) error {
	req := c.Locals("req").(*wallet.DeleteRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*wallet.SimpleResponse, error) {
		return w.client.DeleteCategory(ctx, req)
	})
}

func (w *FinanceCallsParams) GetTransactions(c fiber.Ctx) error {
	req := c.Locals("req").(*wallet.GetTransactionsRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*wallet.GetTransactionsResponse, error) {
		return w.client.GetTransactions(ctx, req)
	})
}

func (w *FinanceCallsParams) CreateTransaction(c fiber.Ctx) error {
	req := c.Locals("req").(*wallet.CreateTransactionRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*wallet.SimpleResponse, error) {
		return w.client.CreateTransaction(ctx, req)
	})
}

func (w *FinanceCallsParams) UpdateTransaction(c fiber.Ctx) error {
	req := c.Locals("req").(*wallet.UpdateTransactionRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*wallet.SimpleResponse, error) {
		return w.client.UpdateTransaction(ctx, req)
	})
}

func (w *FinanceCallsParams) DeleteTransaction(c fiber.Ctx) error {
	req := c.Locals("req").(*wallet.DeleteRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*wallet.SimpleResponse, error) {
		return w.client.DeleteTransaction(ctx, req)
	})
}

func (w *FinanceCallsParams) ExportTransactions(c fiber.Ctx) error {
	req := c.Locals("req").(*wallet.ExportTransactionsRequest)

	res, err := rpc.Call(c, func(ctx context.Context) (*wallet.ExportTransactionsResponse, error) {
		return w.client.ExportTransactions(ctx, req)
	})
	if err != nil {
		return err
	}

	c.Set("Content-Disposition", "attachment; filename="+res.FileName)
	c.Set("Content-Type", res.ContentType)
	return c.Send(res.FileData)
}

func (w *FinanceCallsParams) GetTransactionsUpdateHistory(c fiber.Ctx) error {
	req := c.Locals("req").(*wallet.GetTransactionsUpdateHistoryRequest)

	return rpc.CallWithJSON(c, func(ctx context.Context) (*wallet.GetTransactionsUpdateHistoryResponse, error) {
		return w.client.GetTransactionsUpdateHistory(ctx, req)
	})
}
