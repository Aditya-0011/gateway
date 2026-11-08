package middlewares

import (
	"time"

	"gateway/db"

	"github.com/gofiber/fiber/v3"
)

type SessionCookie struct {
	name   string
	domain string
	secure bool
}

func NewSessionCookie(domain string, secure bool) SessionCookie {
	name := "session"
	if secure {
		name = "__Secure-session"
	}
	return SessionCookie{name: name, domain: domain, secure: secure}
}

func (s SessionCookie) Set(c fiber.Ctx, value string) {
	c.Cookie(s.cookie(value, int(db.SessionIdleTimeout.Seconds())))
}

func (s SessionCookie) Clear(c fiber.Ctx) {
	cookie := s.cookie("", -1)
	cookie.Expires = time.Unix(0, 0)
	c.Cookie(cookie)
}

func (s SessionCookie) Read(c fiber.Ctx) string {
	return c.Cookies(s.name)
}

func (s SessionCookie) cookie(value string, maxAge int) *fiber.Cookie {
	return &fiber.Cookie{
		Name:     s.name,
		Value:    value,
		Path:     "/",
		Domain:   s.domain,
		HTTPOnly: true,
		Secure:   s.secure,
		SameSite: fiber.CookieSameSiteStrictMode,
		MaxAge:   maxAge,
	}
}
