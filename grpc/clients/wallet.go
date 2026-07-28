package clients

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/Aditya-0011/common/contracts/go/wallet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func setupWalletClient() (*grpc.ClientConn, wallet.UserServiceClient, wallet.FinanceServiceClient) {
	addr := os.Getenv("WALLET_SERVICE_ADDR")
	if addr == "" {
		slog.LogAttrs(context.Background(), slog.LevelError, "WALLET_SERVICE_ADDR environment variable is not set")
		os.Exit(1)
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                5 * time.Minute,
			Timeout:             3 * time.Second,
			PermitWithoutStream: false,
		}),
	)
	if err != nil {
		slog.LogAttrs(context.Background(), slog.LevelError, "Failed to connect to wallet service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	walletUserServiceClient := wallet.NewUserServiceClient(conn)
	walletFinanceServiceClient := wallet.NewFinanceServiceClient(conn)

	slog.LogAttrs(context.Background(), slog.LevelInfo, "Connected to wallet service")
	return conn, walletUserServiceClient, walletFinanceServiceClient
}

func closeWalletClient(conn *grpc.ClientConn) error {
	slog.LogAttrs(context.Background(), slog.LevelInfo, "Closing wallet service connection")
	return conn.Close()
}
