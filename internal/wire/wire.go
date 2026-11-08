package wire

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	marshalOptions = protojson.MarshalOptions{
		UseProtoNames:     true,
		UseEnumNumbers:    true,
		EmitDefaultValues: true,
	}
	unmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

func Decode(data []byte, msg proto.Message) error {
	return unmarshalOptions.Unmarshal(data, msg)
}

func Write(c fiber.Ctx, msg proto.Message) error {
	body, err := marshalOptions.Marshal(msg)
	if err != nil {
		slog.LogAttrs(c.Context(), slog.LevelError, "encoding response", slog.Any("error", err))
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	return c.Send(body)
}
