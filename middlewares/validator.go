package middlewares

import (
	"gateway/schema"
	"strings"
	"sync"

	"buf.build/go/protovalidate"
	"github.com/gofiber/fiber/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func trimStrings(msg proto.Message) {
	if msg == nil {
		return
	}
	m := msg.ProtoReflect()
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if !fd.IsList() && !fd.IsMap() {
			if fd.Kind() == protoreflect.StringKind {
				m.Set(fd, protoreflect.ValueOfString(strings.TrimSpace(v.String())))
			} else if fd.Kind() == protoreflect.MessageKind {
				trimStrings(v.Message().Interface())
			}
			return true
		}

		if fd.IsList() {
			list := m.Get(fd).List()
			for i := 0; i < list.Len(); i++ {
				if fd.Kind() == protoreflect.StringKind {
					val := list.Get(i).String()
					list.Set(i, protoreflect.ValueOfString(strings.TrimSpace(val)))
				} else if fd.Kind() == protoreflect.MessageKind {
					trimStrings(list.Get(i).Message().Interface())
				}
			}
		}

		return true
	})
}

func Validate[T any, Ptr interface {
	*T
	proto.Message
}](validator protovalidate.Validator) fiber.Handler {
	pool := sync.Pool{
		New: func() any {
			var msg Ptr = new(T)
			return msg
		},
	}

	return func(c fiber.Ctx) error {
		msg := pool.Get().(Ptr)
		defer func() {
			proto.Reset(msg)
			pool.Put(msg)
		}()

		if len(c.Body()) > 0 {
			if err := c.Bind().JSON(msg); err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Invalid request body format")
			}
		}

		trimStrings(msg)

		if authInfo, ok := c.Locals("auth").(*schema.AuthInfo); ok {
			m := msg.ProtoReflect()
			if fd := m.Descriptor().Fields().ByName("user_id"); fd != nil {
				if fd.Kind() == protoreflect.Int32Kind {
					m.Set(fd, protoreflect.ValueOfInt32(int32(authInfo.UserId)))
				} else if fd.Kind() == protoreflect.Int64Kind {
					m.Set(fd, protoreflect.ValueOfInt64(int64(authInfo.UserId)))
				}
			}
		}

		if err := validator.Validate(msg); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		c.Locals("req", msg)

		return c.Next()
	}
}
