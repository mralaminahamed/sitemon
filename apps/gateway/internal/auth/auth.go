// Package auth guards the /api routes with a shared API key.
package auth

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// APIKey requires X-API-Key (or Authorization: Bearer <key>) to equal key.
func APIKey(key string) echo.MiddlewareFunc {
	want := []byte(key)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			got := c.Request().Header.Get("X-API-Key")
			if got == "" {
				got = strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
			}
			if got == "" {
				// Browsers can't set headers on a WebSocket handshake, so also
				// accept the key as a query parameter.
				got = c.QueryParam("api_key")
			}
			if subtle.ConstantTimeCompare([]byte(got), want) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or missing API key")
			}
			return next(c)
		}
	}
}
