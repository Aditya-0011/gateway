package calls

import (
	"context"
	"gateway/utils"

	"github.com/gofiber/fiber/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Call[T any](c fiber.Ctx, action func(context.Context) (T, error)) (T, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	res, err := action(ctx)
	if err != nil {
		var zero T
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return zero, fiber.NewError(fiber.StatusNotFound, st.Message())
			case codes.Unauthenticated:
				return zero, fiber.NewError(fiber.StatusUnauthorized, st.Message())
			default:
				return zero, fiber.NewError(fiber.StatusInternalServerError, st.Message())
			}
		}
		return zero, fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}

	return res, nil
}

func CallWithJSON[T any](c fiber.Ctx, action func(context.Context) (T, error)) error {
	res, err := Call(c, action)
	if err != nil {
		return err
	}
	return c.JSON(res)
}
