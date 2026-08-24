package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestAPIKey(t *testing.T) {
	e := echo.New()
	mw := APIKey("secret")
	h := mw(func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	tests := []struct {
		name, header, value string
		want                int
	}{
		{"missing", "", "", http.StatusUnauthorized},
		{"wrong", "X-API-Key", "nope", http.StatusUnauthorized},
		{"correct x-api-key", "X-API-Key", "secret", http.StatusOK},
		{"correct bearer", "Authorization", "Bearer secret", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
			if tt.header != "" {
				req.Header.Set(tt.header, tt.value)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			err := h(c)
			code := rec.Code
			if he, ok := err.(*echo.HTTPError); ok {
				code = he.Code
			}
			if code != tt.want {
				t.Errorf("code = %d, want %d", code, tt.want)
			}
		})
	}
}
