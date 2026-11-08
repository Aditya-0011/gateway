package routes

import (
	calls "gateway/grpc/calls/wallet"
	"gateway/grpc/clients"
	"gateway/middlewares"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/wallet"
	"github.com/gofiber/fiber/v3"
)

func walletRoutes(r fiber.Router, cl *clients.Clients, authenticate, access fiber.Handler, v protovalidate.Validator) {
	user := calls.WalletUserCalls(cl.WalletUser)
	finance := calls.FinanceCalls(cl.WalletFinance)
	session := middlewares.RequireSession()

	r.Get("/user/summary", authenticate, access, session, middlewares.Validate[*wallet.SimpleRequest](v), user.GetUserSummary)

	r.Get("/category/list", authenticate, access, session, middlewares.Validate[*wallet.SimpleRequest](v), finance.GetCategories)
	r.Post("/category/add", authenticate, access, session, middlewares.Validate[*wallet.CreateCategoryRequest](v), finance.CreateCategory)
	r.Post("/category/edit", authenticate, access, session, middlewares.Validate[*wallet.UpdateCategoryRequest](v), finance.UpdateCategory)
	r.Post("/category/delete", authenticate, access, session, middlewares.Validate[*wallet.DeleteRequest](v), finance.DeleteCategory)

	r.Query("/transaction/list", authenticate, access, session, middlewares.Validate[*wallet.GetTransactionsRequest](v), finance.GetTransactions)
	r.Post("/transaction/add", authenticate, access, session, middlewares.Validate[*wallet.CreateTransactionRequest](v), finance.CreateTransaction)
	r.Post("/transaction/edit", authenticate, access, session, middlewares.Validate[*wallet.UpdateTransactionRequest](v), finance.UpdateTransaction)
	r.Post("/transaction/delete", authenticate, access, session, middlewares.Validate[*wallet.DeleteRequest](v), finance.DeleteTransaction)
	r.Query("/transaction/export", authenticate, access, session, middlewares.Validate[*wallet.ExportTransactionsRequest](v), finance.ExportTransactions)
	r.Query("/transaction/history", authenticate, access, session, middlewares.Validate[*wallet.GetTransactionsUpdateHistoryRequest](v), finance.GetTransactionsUpdateHistory)
}
