package middlewares

import (
	"strings"

	"gateway/internal/wire"
	"gateway/schema"

	"buf.build/go/protovalidate"
	"github.com/gofiber/fiber/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Validate[M proto.Message](validator protovalidate.Validator) fiber.Handler {
	var zero M
	messageType := zero.ProtoReflect().Type()

	return func(c fiber.Ctx) error {
		msg := messageType.New().Interface().(M)

		if len(c.Body()) > 0 {
			if !strings.HasPrefix(c.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
				return fiber.NewError(fiber.StatusUnsupportedMediaType, "Content-Type must be application/json")
			}
			if err := wire.Decode(c.Body(), msg); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
			}
		}

		trimStrings(msg.ProtoReflect())

		if info, ok := c.Locals("auth").(*schema.AuthInfo); ok {
			setUserID(msg.ProtoReflect(), info.UserID)
		}

		if err := validator.Validate(msg); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		c.Locals("req", msg)
		return c.Next()
	}
}

func setUserID(m protoreflect.Message, userID int) {
	fd := m.Descriptor().Fields().ByName("user_id")
	if fd == nil {
		return
	}
	switch fd.Kind() {
	case protoreflect.Int32Kind:
		m.Set(fd, protoreflect.ValueOfInt32(int32(userID)))
	case protoreflect.Int64Kind:
		m.Set(fd, protoreflect.ValueOfInt64(int64(userID)))
	}
}

func trimStrings(m protoreflect.Message) {
	type update struct {
		fd  protoreflect.FieldDescriptor
		val protoreflect.Value
	}
	var updates []update

	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch {
		case fd.IsMap():
		case fd.IsList():
			list := v.List()
			for i := 0; i < list.Len(); i++ {
				switch fd.Kind() {
				case protoreflect.StringKind:
					list.Set(i, protoreflect.ValueOfString(strings.TrimSpace(list.Get(i).String())))
				case protoreflect.MessageKind:
					trimStrings(list.Get(i).Message())
				}
			}
		case fd.Kind() == protoreflect.StringKind && !isSecretField(fd):
			if trimmed := strings.TrimSpace(v.String()); trimmed != v.String() {
				updates = append(updates, update{fd, protoreflect.ValueOfString(trimmed)})
			}
		case fd.Kind() == protoreflect.MessageKind:
			trimStrings(v.Message())
		}
		return true
	})

	for _, u := range updates {
		m.Set(u.fd, u.val)
	}
}

func isSecretField(fd protoreflect.FieldDescriptor) bool {
	return strings.HasSuffix(string(fd.Name()), "password")
}
