package clients

import (
	"context"
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
		slog.LogAttrs(context.Background(), slog.LevelError, "AUTH_SERVICE_ADDR environment variable is not set")
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
		slog.LogAttrs(context.Background(), slog.LevelError, "Failed to connect to auth service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	authClient := auth.NewAuthServiceClient(conn)

	slog.LogAttrs(context.Background(), slog.LevelInfo, "Connected to auth service")
	return conn, authClient
}

func closeAuthClient(conn *grpc.ClientConn) error {
	slog.LogAttrs(context.Background(), slog.LevelInfo, "Closing auth service connection")
	return conn.Close()
}
