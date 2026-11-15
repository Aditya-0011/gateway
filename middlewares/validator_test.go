package middlewares

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gateway/schema"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/gofiber/fiber/v3"
)

func TestValidateOverwritesUserID(t *testing.T) {
	v, err := protovalidate.New()
	if err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	var got int32
	app.Post("/", func(c fiber.Ctx) error {
		c.Locals("auth", &schema.AuthInfo{Identity: schema.Identity{UserID: 7}, SessionID: "sid"})
		return c.Next()
	}, Validate[*auth.UserRequest](v), func(c fiber.Ctx) error {
		got = c.Locals("req").(*auth.UserRequest).GetUserId()
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest(fiber.MethodPost, "/", strings.NewReader(`{"user_id": 99}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("status = %d, want 204", res.StatusCode)
	}
	if got != 7 {
		t.Fatalf("user_id = %d, want 7 from the authenticated caller", got)
	}
}

func TestValidateRejectsNonJSONBody(t *testing.T) {
	v, err := protovalidate.New()
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Post("/", Validate[*auth.UserRequest](v), func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	req := httptest.NewRequest(fiber.MethodPost, "/", strings.NewReader(`{"user_id": 1}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMETextPlain)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want 415", res.StatusCode)
	}
}
