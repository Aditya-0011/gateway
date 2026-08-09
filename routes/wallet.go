package routes

import (
	calls "gateway/grpc/calls/wallet"
	"gateway/middlewares"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/wallet"
	"github.com/gofiber/fiber/v3"
)

func walletRouter(router fiber.Router, requireHost fiber.Handler, authMiddleware fiber.Handler, requireSession fiber.Handler, userClient wallet.UserServiceClient, financeClient wallet.FinanceServiceClient, validator protovalidate.Validator) {
	walletUserCall := calls.WalletUserCalls(userClient, validator)
	financeCall := calls.FinanceCalls(financeClient, validator)

	// User
	router.Get("/user/summary", requireHost, authMiddleware, requireSession, middlewares.Validate[wallet.SimpleRequest](validator), walletUserCall.GetUserSummary)

	// Category
	router.Get("/category/list", requireHost, authMiddleware, requireSession, middlewares.Validate[wallet.SimpleRequest](validator), financeCall.GetCategories)
	router.Post("/category/add", requireHost, authMiddleware, requireSession, middlewares.Validate[wallet.CreateCategoryRequest](validator), financeCall.CreateCategory)
	router.Post("/category/edit", requireHost, authMiddleware, requireSession, middlewares.Validate[wallet.UpdateCategoryRequest](validator), financeCall.UpdateCategory)
	router.Post("/category/delete", requireHost, authMiddleware, requireSession, middlewares.Validate[wallet.DeleteRequest](validator), financeCall.DeleteCategory)

	// Transaction
	router.Query("/transaction/list", requireHost, authMiddleware, requireSession, middlewares.Validate[wallet.GetTransactionsRequest](validator), financeCall.GetTransactions)
	router.Post("/transaction/add", requireHost, authMiddleware, requireSession, middlewares.Validate[wallet.CreateTransactionRequest](validator), financeCall.CreateTransaction)
	router.Post("/transaction/edit", requireHost, authMiddleware, requireSession, middlewares.Validate[wallet.UpdateTransactionRequest](validator), financeCall.UpdateTransaction)
	router.Post("/transaction/delete", requireHost, authMiddleware, requireSession, middlewares.Validate[wallet.DeleteRequest](validator), financeCall.DeleteTransaction)
	router.Query("/transaction/export", requireHost, authMiddleware, requireSession, middlewares.Validate[wallet.ExportTransactionsRequest](validator), financeCall.ExportTransactions)
	router.Query("/transaction/history", requireHost, authMiddleware, requireSession, middlewares.Validate[wallet.GetTransactionsUpdateHistoryRequest](validator), financeCall.GetTransactionsUpdateHistory)
}
