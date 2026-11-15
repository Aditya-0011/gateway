package middlewares

import (
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestRateLimitBlocksAfterLimit(t *testing.T) {
	store := testStore(t)
	limit := RateLimit{
		Name:    "test-" + randomSuffix(),
		Limit:   3,
		Window:  time.Minute,
		Message: "slow down",
		Skip:    func(c fiber.Ctx) bool { return c.Get("X-Skip") != "" },
	}
	app := fiber.New()
	app.Get("/", limit.Handler(store), func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	do := func(skip bool) int {
		req := httptest.NewRequest(fiber.MethodGet, "/", nil)
		if skip {
			req.Header.Set("X-Skip", "1")
		}
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode == fiber.StatusTooManyRequests {
			if after, err := strconv.Atoi(res.Header.Get(fiber.HeaderRetryAfter)); err != nil || after <= 0 {
				t.Fatalf("Retry-After = %q", res.Header.Get(fiber.HeaderRetryAfter))
			}
		}
		return res.StatusCode
	}

	for i := range 3 {
		if got := do(false); got != fiber.StatusNoContent {
			t.Fatalf("request %d: status = %d, want 204", i+1, got)
		}
	}
	if got := do(false); got != fiber.StatusTooManyRequests {
		t.Fatalf("request 4: status = %d, want 429", got)
	}
	if got := do(true); got != fiber.StatusNoContent {
		t.Fatalf("skipped request: status = %d, want 204", got)
	}
}
