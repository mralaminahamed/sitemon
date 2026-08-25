package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/readmodel"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/service"
)

func newHandler() *Handler { return New(service.New(nil), readmodel.New(nil, nil)) }

func TestHealth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := newHandler().Health(c); err != nil {
		t.Fatalf("Health: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "gateway") {
		t.Errorf("body missing service name: %s", rec.Body.String())
	}
}

func TestChecksValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{"empty urls", `{"urls":[]}`, http.StatusBadRequest},
		{"invalid json", `{`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/checks", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := newHandler().Checks(c); err != nil {
				t.Fatalf("Checks: %v", err)
			}
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestChecksTooManyURLs(t *testing.T) {
	var b strings.Builder
	b.WriteString(`{"urls":[`)
	for i := 0; i <= maxCheckURLs; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(`"https://x.com"`)
	}
	b.WriteString(`]}`)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/checks", strings.NewReader(b.String()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := newHandler().Checks(c); err != nil {
		t.Fatalf("Checks: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestDeleteMonitorRequiresURL(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/monitors", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := newHandler().DeleteMonitor(c); err != nil {
		t.Fatalf("DeleteMonitor: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestSSLRequiresURL(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/ssl", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := newHandler().SSL(c); err != nil {
		t.Fatalf("SSL: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestLoadTestRequiresURL(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/loadtest", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := newHandler().LoadTest(c); err != nil {
		t.Fatalf("LoadTest: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}
