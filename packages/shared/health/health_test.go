package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func run(h http.HandlerFunc) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec
}

func TestLiveness(t *testing.T) {
	rec := run(LivenessHandler("gw"))
	if rec.Code != http.StatusOK {
		t.Errorf("code = %d, want 200", rec.Code)
	}
}

func TestReadyAllOK(t *testing.T) {
	rec := run(ReadyHandler("gw", Check{"nats", func(context.Context) error { return nil }}))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"ready"`) {
		t.Errorf("want 200 ready, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestReadyOneFails(t *testing.T) {
	rec := run(ReadyHandler("gw",
		Check{"nats", func(context.Context) error { return nil }},
		Check{"mongo", func(context.Context) error { return errors.New("timeout") }},
	))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("code = %d, want 503", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "not ready") {
		t.Errorf("body missing 'not ready': %s", rec.Body.String())
	}
}
