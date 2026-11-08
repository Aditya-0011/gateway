package calls

import (
	"context"
	"log/slog"
	"time"

	"gateway/internal/wire"

	"github.com/Aditya-0011/common/contracts/go/grpcx"
	"github.com/gofiber/fiber/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const Timeout = 9500 * time.Millisecond

func Call[T proto.Message](c fiber.Ctx, action func(context.Context) (T, error)) (T, error) {
	ctx, cancel := context.WithTimeout(c.Context(), Timeout)
	defer cancel()

	res, err := action(ctx)
	if err != nil {
		var zero T
		return zero, toHTTPError(c, err)
	}
	return res, nil
}

func CallWithJSON[T proto.Message](c fiber.Ctx, action func(context.Context) (T, error)) error {
	res, err := Call(c, action)
	if err != nil {
		return err
	}
	return wire.Write(c, res)
}

func toHTTPError(c fiber.Ctx, err error) error {
	st := status.Convert(err)
	msg := st.Message()

	switch st.Code() {
	case codes.InvalidArgument, codes.OutOfRange:
		return fiber.NewError(fiber.StatusBadRequest, msg)
	case codes.Unauthenticated:
		return fiber.NewError(fiber.StatusUnauthorized, msg)
	case codes.NotFound:
		return fiber.NewError(fiber.StatusNotFound, msg)
	case codes.AlreadyExists, codes.Aborted:
		return fiber.NewError(fiber.StatusConflict, msg)
	case codes.FailedPrecondition:
		return fiber.NewError(fiber.StatusPreconditionFailed, msg)
	case codes.ResourceExhausted:
		return fiber.NewError(fiber.StatusTooManyRequests, msg)
	case codes.DeadlineExceeded:
		return fiber.NewError(fiber.StatusGatewayTimeout, "Upstream timed out")
	case codes.Unavailable, codes.Canceled:
		return fiber.NewError(fiber.StatusServiceUnavailable, "Service unavailable")
	case codes.PermissionDenied:
		if grpcx.IsServiceCredentialsError(err) {
			slog.LogAttrs(c.Context(), slog.LevelError, "upstream rejected gateway credentials", slog.String("error", msg))
			return fiber.NewError(fiber.StatusBadGateway, "Bad gateway")
		}
		return fiber.NewError(fiber.StatusForbidden, msg)
	default:
		slog.LogAttrs(c.Context(), slog.LevelError, "upstream error",
			slog.String("code", st.Code().String()), slog.String("error", msg))
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}
}
