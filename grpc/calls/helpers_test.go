package calls

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestToHTTPError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"invalid argument", status.Error(codes.InvalidArgument, "bad"), fiber.StatusBadRequest},
		{"not found", status.Error(codes.NotFound, "missing"), fiber.StatusNotFound},
		{"rate limited upstream", status.Error(codes.ResourceExhausted, "slow down"), fiber.StatusTooManyRequests},
		{"response too large", status.Error(codes.ResourceExhausted, "grpc: received message larger than max (5000000 vs. 4194304)"), fiber.StatusBadGateway},
		{"unavailable", status.Error(codes.Unavailable, "down"), fiber.StatusServiceUnavailable},
		{"unknown", status.Error(codes.Internal, "boom"), fiber.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c fiber.Ctx) error { return toHTTPError(c, tt.err) })
			res, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", res.StatusCode, tt.want)
			}
		})
	}
}
