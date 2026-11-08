package clients

import (
	"errors"
	"fmt"

	"gateway/internal/config"

	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/Aditya-0011/common/contracts/go/grpcx"
	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/Aditya-0011/common/contracts/go/wallet"
	"google.golang.org/grpc"
)

type Clients struct {
	conns []*grpc.ClientConn

	Auth             auth.AuthServiceClient
	ManagerUser      manager.UserServiceClient
	ManagerPortfolio manager.PortfolioServiceClient
	WalletUser       wallet.UserServiceClient
	WalletFinance    wallet.FinanceServiceClient
}

func New(cfg config.Config) (*Clients, error) {
	c := &Clients{}

	dial := func(name string, sc grpcx.ClientConfig) (*grpc.ClientConn, error) {
		conn, err := grpcx.Dial(sc)
		if err != nil {
			return nil, fmt.Errorf("%s service: %w", name, err)
		}
		c.conns = append(c.conns, conn)
		return conn, nil
	}

	authConn, err := dial("auth", cfg.AuthService)
	if err != nil {
		return nil, errors.Join(err, c.Close())
	}
	managerConn, err := dial("manager", cfg.ManagerService)
	if err != nil {
		return nil, errors.Join(err, c.Close())
	}
	walletConn, err := dial("wallet", cfg.WalletService)
	if err != nil {
		return nil, errors.Join(err, c.Close())
	}

	c.Auth = auth.NewAuthServiceClient(authConn)
	c.ManagerUser = manager.NewUserServiceClient(managerConn)
	c.ManagerPortfolio = manager.NewPortfolioServiceClient(managerConn)
	c.WalletUser = wallet.NewUserServiceClient(walletConn)
	c.WalletFinance = wallet.NewFinanceServiceClient(walletConn)
	return c, nil
}

func (c *Clients) Close() error {
	var errs []error
	for _, conn := range c.conns {
		errs = append(errs, conn.Close())
	}
	return errors.Join(errs...)
}
