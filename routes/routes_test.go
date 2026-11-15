package routes

import (
	"context"
	"math/rand/v2"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"gateway/db"
	"gateway/internal/config"
	"gateway/middlewares"
	"gateway/schema"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/gofiber/fiber/v3"
)

func testStore(t *testing.T) *db.Store {
	t.Helper()
	url := os.Getenv("GATEWAY_TEST_REDIS_URL")
	if url == "" {
		t.Skip("GATEWAY_TEST_REDIS_URL is not set")
	}
	store, err := db.Connect(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestVisitorSubject(t *testing.T) {
	serverKey := &schema.AuthInfo{APIKey: true, KeyOrigins: []string{}}
	browserKey := &schema.AuthInfo{APIKey: true, KeyOrigins: []string{"https://portfolio.example"}}
	session := &schema.AuthInfo{SessionID: "sid"}

	tests := []struct {
		name      string
		info      *schema.AuthInfo
		visitorIP string
		origin    string
		want      string
	}{
		{"server key uses visitor ip", serverKey, "203.0.113.9", "", "203.0.113.9"},
		{"server key masks ipv6 to /64", serverKey, "2001:db8:1:2:3:4:5:6", "", "2001:db8:1:2::/64"},
		{"server key with invalid header", serverKey, "nope", "", ""},
		{"server key from a browser", serverKey, "203.0.113.9", "https://portfolio.example", ""},
		{"browser key ignores header", browserKey, "203.0.113.9", "", ""},
		{"session ignores header", session, "203.0.113.9", "", ""},
		{"unauthenticated ignores header", nil, "203.0.113.9", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got, ip string
			app := fiber.New()
			app.Get("/", func(c fiber.Ctx) error {
				if tt.info != nil {
					c.Locals("auth", tt.info)
				}
				got, ip = visitorSubject(c), c.IP()
				return c.SendStatus(fiber.StatusNoContent)
			})
			req := httptest.NewRequest(fiber.MethodGet, "/", nil)
			req.Header.Set(visitorIPHeader, tt.visitorIP)
			if tt.origin != "" {
				req.Header.Set(fiber.HeaderOrigin, tt.origin)
			}
			if _, err := app.Test(req); err != nil {
				t.Fatal(err)
			}
			want := tt.want
			if want == "" {
				want = ip
			}
			if got != want {
				t.Fatalf("visitorSubject() = %q, want %q", got, want)
			}
		})
	}
}

func TestCORSLanes(t *testing.T) {
	store := testStore(t)
	const portfolioOrigin = "https://portfolio.example"
	if err := store.CacheKeyOrigins(context.Background(), []string{portfolioOrigin}); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{APIDomain: "api.example.com", AllowedOrigins: []string{"https://console.example.com"}}

	app := fiber.New()
	app.Use(CORS(cfg, store, nil))
	app.All("/*", func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	tests := []struct {
		name       string
		host       string
		path       string
		origin     string
		wantOrigin string
		wantCreds  bool
	}{
		{"key origin on manager portfolio path", "manager.api.example.com", "/user/details", portfolioOrigin, portfolioOrigin, false},
		{"key origin on manager owner path", "manager.api.example.com", "/user/edit", portfolioOrigin, "", false},
		{"key origin on another host", "auth.api.example.com", "/user/details", portfolioOrigin, "", false},
		{"unknown origin on portfolio path", "manager.api.example.com", "/user/details", "https://evil.example", "", false},
		{"first party origin", "auth.api.example.com", "/session", "https://console.example.com", "https://console.example.com", true},
		{"first party on portfolio path", "manager.api.example.com", "/user/details", "https://console.example.com", "https://console.example.com", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodOptions, "http://"+tt.host+tt.path, nil)
			req.Header.Set(fiber.HeaderOrigin, tt.origin)
			req.Header.Set(fiber.HeaderAccessControlRequestMethod, fiber.MethodGet)
			res, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if got := res.Header.Get(fiber.HeaderAccessControlAllowOrigin); got != tt.wantOrigin {
				t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantOrigin)
			}
			if got := res.Header.Get(fiber.HeaderAccessControlAllowCredentials) == "true"; got != tt.wantCreds {
				t.Fatalf("Access-Control-Allow-Credentials = %v, want %v", got, tt.wantCreds)
			}
		})
	}
}

func TestAccountLimiter(t *testing.T) {
	store := testStore(t)
	v, err := protovalidate.New()
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Post("/login", middlewares.Validate[*auth.LoginRequest](v), accountLimiter[*auth.LoginRequest](store),
		func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	email := "limit-" + strings.ReplaceAll(t.Name(), "/", "-") + "-" + randomToken() + "@example.com"
	do := func(addr string) int {
		req := httptest.NewRequest(fiber.MethodPost, "/login", strings.NewReader(`{"email":"`+addr+`","password":"password123"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		return res.StatusCode
	}
	for i := range 10 {
		if got := do(email); got != fiber.StatusNoContent {
			t.Fatalf("attempt %d: status = %d, want 204", i+1, got)
		}
	}
	if got := do(strings.ToUpper(email)); got != fiber.StatusTooManyRequests {
		t.Fatalf("attempt 11 with different case: status = %d, want 429", got)
	}
}

func TestHealthExactPath(t *testing.T) {
	store := testStore(t)
	app := fiber.New()
	app.Get("/healthz", Health(store))
	app.Use(notFoundHandler)

	tests := []struct {
		path string
		want int
	}{
		{"/healthz", fiber.StatusOK},
		{"/healthz?x=1", fiber.StatusOK},
		{"/healthz/", fiber.StatusNotFound},
		{"/HEALTHZ", fiber.StatusNotFound},
		{"/Healthz", fiber.StatusNotFound},
		{"//healthz", fiber.StatusNotFound},
		{"/./healthz", fiber.StatusNotFound},
		{"/%68ealthz", fiber.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			res, err := app.Test(httptest.NewRequest(fiber.MethodGet, tt.path, nil))
			if err != nil {
				t.Fatal(err)
			}
			if res.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", res.StatusCode, tt.want)
			}
		})
	}
}

func randomToken() string {
	b := make([]byte, 12)
	for i := range b {
		b[i] = "abcdefghijklmnopqrstuvwxyz"[rand.IntN(26)]
	}
	return string(b)
}
