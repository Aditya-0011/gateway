package clients

import (
	"log/slog"
	"os"
	"time"

	"github.com/Aditya-0011/common/contracts/go/manager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func setupManagerClient() (*grpc.ClientConn, manager.UserServiceClient, manager.PortfolioServiceClient) {
	addr := os.Getenv("MANAGER_SERVICE_ADDR")
	if addr == "" {
		slog.Error("MANAGER_SERVICE_ADDR environment variable is not set")
		os.Exit(1)
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                5 * time.Minute,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		slog.Error("Failed to connect to manager service", "error", err)
		os.Exit(1)
	}

	managerUserServiceClient := manager.NewUserServiceClient(conn)
	managerPortfolioServiceClient := manager.NewPortfolioServiceClient(conn)

	slog.Info("Connected to manager service")
	return conn, managerUserServiceClient, managerPortfolioServiceClient
}

func closeManagerClient(conn *grpc.ClientConn) error {
	slog.Info("Closing manager service connection")
	return conn.Close()
}
