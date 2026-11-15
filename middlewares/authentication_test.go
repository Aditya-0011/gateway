package middlewares

import (
	"context"
	"math/rand/v2"
	"net/http/httptest"
	"os"
	"testing"

	"gateway/db"
	"gateway/schema"

	"github.com/Aditya-0011/common/contracts/go/auth"
	"github.com/gofiber/fiber/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const firstPartyOrigin = "https://console.example.com"

type fakeAuth struct {
	auth.AuthServiceClient
	keys map[string]*auth.ApiKeyIdentity
}

func (f fakeAuth) ValidateKey(_ context.Context, in *auth.ValidateKeyRequest, _ ...grpc.CallOption) (*auth.ApiKeyIdentity, error) {
	if id, ok := f.keys[in.GetKey()]; ok {
		return id, nil
	}
	return nil, status.Error(codes.NotFound, "invalid key")
}

func (f fakeAuth) GetIdentity(_ context.Context, in *auth.UserRequest, _ ...grpc.CallOption) (*auth.Identity, error) {
	return &auth.Identity{UserId: in.GetUserId(), Email: "user@example.com", Role: auth.Role_ROLE_TENANT, Projects: []string{"manager"}}, nil
}

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

func testUserID() int { return 1_000_000 + rand.IntN(1_000_000_000) }

type authApp struct {
	app    *fiber.App
	cookie string
	seen   *schema.AuthInfo
}

func newAuthApp(t *testing.T, keys map[string]*auth.ApiKeyIdentity) *authApp {
	t.Helper()
	store := testStore(t)
	cookie := NewSessionCookie("", false)
	authn := NewAuthenticator(store, fakeAuth{keys: keys}, cookie, []string{firstPartyOrigin})

	value, _, err := store.CreateSession(context.Background(), testUserID(), db.SessionMeta{IP: "127.0.0.1"})
	if err != nil {
		t.Fatal(err)
	}

	a := &authApp{app: fiber.New(), cookie: value}
	handler := func(c fiber.Ctx) error {
		a.seen = c.Locals("auth").(*schema.AuthInfo)
		return c.SendStatus(fiber.StatusNoContent)
	}
	a.app.Get("/", authn.Required(), handler)
	a.app.Post("/", authn.Required(), handler)
	return a
}

func (a *authApp) do(t *testing.T, method string, headers map[string]string, withCookie bool) int {
	t.Helper()
	a.seen = nil
	req := httptest.NewRequest(method, "/", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if withCookie {
		req.Header.Set(fiber.HeaderCookie, "session="+a.cookie)
	}
	res, err := a.app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return res.StatusCode
}

func TestAPIKeyNeverFallsBackToCookie(t *testing.T) {
	a := newAuthApp(t, nil)
	if got := a.do(t, fiber.MethodGet, map[string]string{APIKeyHeader: "not-a-real-key"}, true); got != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", got)
	}
	if got := a.do(t, fiber.MethodGet, nil, true); got != fiber.StatusNoContent {
		t.Fatalf("cookie alone: status = %d, want 204", got)
	}
}

func TestSessionOriginCheck(t *testing.T) {
	a := newAuthApp(t, nil)
	tests := []struct {
		name   string
		method string
		origin string
		want   int
	}{
		{"post from first party", fiber.MethodPost, firstPartyOrigin, fiber.StatusNoContent},
		{"post from foreign origin", fiber.MethodPost, "https://evil.example", fiber.StatusForbidden},
		{"post without origin", fiber.MethodPost, "", fiber.StatusForbidden},
		{"get without origin", fiber.MethodGet, "", fiber.StatusNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{}
			if tt.origin != "" {
				headers[fiber.HeaderOrigin] = tt.origin
			}
			if got := a.do(t, tt.method, headers, true); got != tt.want {
				t.Fatalf("status = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestAPIKeyOrigins(t *testing.T) {
	serverKey := "server-" + randomSuffix()
	browserKey := "browser-" + randomSuffix()
	a := newAuthApp(t, map[string]*auth.ApiKeyIdentity{
		serverKey:  {Identity: &auth.Identity{UserId: int32(testUserID()), Projects: []string{"manager"}}},
		browserKey: {Identity: &auth.Identity{UserId: int32(testUserID()), Projects: []string{"manager"}}, Origins: []string{"https://portfolio.example"}},
	})

	if got := a.do(t, fiber.MethodGet, map[string]string{APIKeyHeader: serverKey}, false); got != fiber.StatusNoContent {
		t.Fatalf("server key: status = %d, want 204", got)
	}
	if !a.seen.IsServerKey() {
		t.Fatal("server key without origins should be a server key")
	}
	if got := a.do(t, fiber.MethodGet, map[string]string{APIKeyHeader: serverKey}, false); got != fiber.StatusNoContent || !a.seen.IsServerKey() {
		t.Fatalf("cached server key: status = %d", got)
	}

	if got := a.do(t, fiber.MethodGet, map[string]string{APIKeyHeader: browserKey, fiber.HeaderOrigin: "https://other.example"}, false); got != fiber.StatusForbidden {
		t.Fatalf("browser key from foreign origin: status = %d, want 403", got)
	}
	if got := a.do(t, fiber.MethodGet, map[string]string{APIKeyHeader: browserKey, fiber.HeaderOrigin: "https://portfolio.example"}, false); got != fiber.StatusNoContent {
		t.Fatalf("browser key from its origin: status = %d, want 204", got)
	}
	if a.seen.IsServerKey() {
		t.Fatal("key with origins must not count as a server key")
	}
}

func randomSuffix() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 24)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}
	return string(b)
}
