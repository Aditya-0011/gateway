package clients

import (
	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/Aditya-0011/common/contracts/go/wallet"
	"google.golang.org/grpc"
)

type ClientParams struct {
	authConn   *grpc.ClientConn
	AuthClient auth.AuthServiceClient

	managerConn            *grpc.ClientConn
	ManagerUserClient      manager.UserServiceClient
	ManagerPortfolioClient manager.PortfolioServiceClient

	walletConn          *grpc.ClientConn
	WalletUserClient    wallet.UserServiceClient
	WalletFinanceClient wallet.FinanceServiceClient
}

func Setup() *ClientParams {
	authConn, authClient := setupAuthClient()
	managerConn, managerUserServiceClient, managerPortfolioServiceClient := setupManagerClient()
	walletConn, walletUserServiceClient, walletFinanceServiceClient := setupWalletClient()

	return &ClientParams{
		authConn:               authConn,
		AuthClient:             authClient,
		managerConn:            managerConn,
		ManagerUserClient:      managerUserServiceClient,
		ManagerPortfolioClient: managerPortfolioServiceClient,
		walletConn:             walletConn,
		WalletUserClient:       walletUserServiceClient,
		WalletFinanceClient:    walletFinanceServiceClient,
	}
}

func (cp *ClientParams) Close() {
	closeAuthClient(cp.authConn)
	closeManagerClient(cp.managerConn)
	closeWalletClient(cp.walletConn)
}
