package routes

import (
	calls "gateway/grpc/calls/wallet"
	"gateway/middlewares"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/wallet"
	"github.com/gofiber/fiber/v3"
)

func walletRouter(router fiber.Router, userClient wallet.UserServiceClient, financeClient wallet.FinanceServiceClient, validator protovalidate.Validator) {
	walletUserCall := calls.WalletUserCalls(userClient, validator)
	financeCall := calls.FinanceCalls(financeClient, validator)

	userApi := router.Group("/user")
	{
		userApi.Get("/summary", middlewares.Validate[wallet.SimpleRequest](validator), walletUserCall.GetUserSummary)
	}

	categoryApi := router.Group("/category")
	{
		categoryApi.Get("/list", middlewares.Validate[wallet.SimpleRequest](validator), financeCall.GetCategories)
		categoryApi.Post("/add", middlewares.Validate[wallet.CreateCategoryRequest](validator), financeCall.CreateCategory)
		categoryApi.Post("/edit", middlewares.Validate[wallet.UpdateCategoryRequest](validator), financeCall.UpdateCategory)
		categoryApi.Post("/delete", middlewares.Validate[wallet.DeleteRequest](validator), financeCall.DeleteCategory)
	}

	transactionApi := router.Group("/transaction")
	{
		transactionApi.Query("/list", middlewares.Validate[wallet.GetTransactionsRequest](validator), financeCall.GetTransactions)
		transactionApi.Post("/add", middlewares.Validate[wallet.CreateTransactionRequest](validator), financeCall.CreateTransaction)
		transactionApi.Post("/edit", middlewares.Validate[wallet.UpdateTransactionRequest](validator), financeCall.UpdateTransaction)
		transactionApi.Post("/delete", middlewares.Validate[wallet.DeleteRequest](validator), financeCall.DeleteTransaction)

		transactionApi.Query("/history", middlewares.Validate[wallet.GetTransactionsUpdateHistoryRequest](validator), financeCall.GetTransactionsUpdateHistory)
	}
}
