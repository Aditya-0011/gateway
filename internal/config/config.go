package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Aditya-0011/common/contracts/go/grpcx"
)

var devOrigins = []string{
	"http://localhost:5173", "http://127.0.0.1:5173",
	"http://localhost:5174", "http://127.0.0.1:5174",
	"http://localhost:5175", "http://127.0.0.1:5175",
}

type Config struct {
	Dev            bool
	Port           string
	Domain         string
	APIDomain      string
	AllowedOrigins []string
	RedisURL       string

	AuthService    grpcx.ClientConfig
	ManagerService grpcx.ClientConfig
	WalletService  grpcx.ClientConfig
}

func Load() (Config, error) {
	var errs []error

	cfg := Config{
		Dev:      os.Getenv("DEVELOPMENT") == "true",
		Port:     envOr("PORT", "3000"),
		Domain:   strings.TrimPrefix(strings.TrimSpace(os.Getenv("DOMAIN")), "."),
		RedisURL: os.Getenv("REDIS_URL"),
	}

	if cfg.RedisURL == "" {
		errs = append(errs, errors.New("REDIS_URL is not set"))
	}
	if cfg.Domain == "" && !cfg.Dev {
		errs = append(errs, errors.New("DOMAIN is not set"))
	}
	cfg.APIDomain = strings.TrimPrefix(strings.TrimSpace(os.Getenv("API_DOMAIN")), ".")
	if cfg.APIDomain == "" && cfg.Domain != "" {
		cfg.APIDomain = "api." + cfg.Domain
	}

	cfg.AllowedOrigins = splitList(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if len(cfg.AllowedOrigins) == 0 {
		if cfg.Dev {
			cfg.AllowedOrigins = devOrigins
		} else {
			errs = append(errs, errors.New("CORS_ALLOWED_ORIGINS is not set"))
		}
	}
	for _, o := range cfg.AllowedOrigins {
		if !cfg.Dev && !strings.HasPrefix(o, "https://") {
			errs = append(errs, fmt.Errorf("CORS origin %q must use https", o))
		}
	}

	token, err := grpcx.TokenFromEnv()
	if err != nil {
		errs = append(errs, err)
	}
	ca := os.Getenv("INTERNAL_TLS_CA")

	service := func(env string) grpcx.ClientConfig {
		addr := os.Getenv(env)
		if addr == "" {
			errs = append(errs, fmt.Errorf("%s is not set", env))
		}
		return grpcx.ClientConfig{Addr: addr, AuthToken: token, TLSCAFile: ca}
	}
	cfg.AuthService = service("AUTH_SERVICE_ADDR")
	cfg.ManagerService = service("MANAGER_SERVICE_ADDR")
	cfg.WalletService = service("WALLET_SERVICE_ADDR")

	return cfg, errors.Join(errs...)
}

func (c Config) Host(service string) string {
	return service + "." + c.APIDomain
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitList(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
