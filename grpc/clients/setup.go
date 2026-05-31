package clients

import (
	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/Aditya-0011/common/contracts/go/manager"
	"google.golang.org/grpc"
)

type ClientParams struct {
	authConn   *grpc.ClientConn
	AuthClient auth.AuthServiceClient

	managerConn            *grpc.ClientConn
	ManagerUserClient      manager.UserServiceClient
	ManagerPortfolioClient manager.PortfolioServiceClient
}

func Setup() *ClientParams {
	authConn, authClient := setupAuthClient()
	managerConn, managerUserServiceClient, managerPortfolioServiceClient := setupManagerClient()

	return &ClientParams{
		authConn:               authConn,
		AuthClient:             authClient,
		managerConn:            managerConn,
		ManagerUserClient:      managerUserServiceClient,
		ManagerPortfolioClient: managerPortfolioServiceClient,
	}
}

func (cp *ClientParams) Close() {
	closeAuthClient(cp.authConn)
	closeManagerClient(cp.managerConn)
}
