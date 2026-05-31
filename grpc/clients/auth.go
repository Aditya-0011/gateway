package clients

import (
	"log/slog"
	"os"
	"time"

	"github.com/Aditya-0011/common/contracts/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func setupAuthClient() (*grpc.ClientConn, auth.AuthServiceClient) {
	addr := os.Getenv("AUTH_SERVICE_ADDR")
	if addr == "" {
		slog.Error("AUTH_SERVICE_ADDR environment variable is not set")
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
		slog.Error("Failed to connect to auth service", "error", err)
		os.Exit(1)
	}

	authClient := auth.NewAuthServiceClient(conn)

	slog.Info("Connected to auth service")
	return conn, authClient
}

func closeAuthClient(conn *grpc.ClientConn) error {
	slog.Info("Closing auth service connection")
	return conn.Close()
}
