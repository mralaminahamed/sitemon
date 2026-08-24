// Package handler holds the Echo HTTP handlers. Handlers validate input, call
// the service layer, and shape the JSON response — no business logic here.
package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/dto"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/service"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/statuscache"
	"github.com/mralaminahamed/sitemon/packages/shared/loadtest"
)

type Handler struct {
	svc   *service.Service
	cache *statuscache.Cache
}

func New(svc *service.Service, cache *statuscache.Cache) *Handler {
	return &Handler{svc: svc, cache: cache}
}

// Health is the liveness probe. Kept at the root path for compose/k8s.
func (h *Handler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"status": "ok", "service": "gateway"})
}

// Root is a small banner listing the API.
func (h *Handler) Root(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{
		"service":     "sitemon-gateway",
		"distributed": h.svc.Distributed(),
		"endpoints": []string{
			"GET  /health",
			"GET  /api/status",
			"POST /api/checks",
			"GET  /api/ssl?url=",
			"POST /api/loadtest",
		},
	})
}

// Status returns the latest health result per URL from the event-driven cache.
func (h *Handler) Status(c echo.Context) error {
	if h.cache == nil {
		return c.JSON(http.StatusOK, echo.Map{"results": []any{}})
	}
	return c.JSON(http.StatusOK, echo.Map{"results": h.cache.Snapshot()})
}

// Checks runs health checks for the requested URLs.
func (h *Handler) Checks(c echo.Context) error {
	var req dto.CheckRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid JSON body")
	}
	if len(req.URLs) == 0 {
		return badRequest(c, "at least one url is required")
	}

	timeout := msOr(req.TimeoutMs, 10*time.Second)
	results := h.svc.CheckURLs(req.URLs, timeout, req.BypassCloudflare)
	return c.JSON(http.StatusOK, echo.Map{"results": results})
}

// SSL returns certificate details for ?url=.
func (h *Handler) SSL(c echo.Context) error {
	url := c.QueryParam("url")
	if url == "" {
		return badRequest(c, "url query parameter is required")
	}

	timeout := 10 * time.Second
	info, err := h.svc.SSL(url, timeout)
	if err != nil {
		return c.JSON(http.StatusBadGateway, dto.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, info)
}

// LoadTest runs a load test and returns aggregate stats.
func (h *Handler) LoadTest(c echo.Context) error {
	var req dto.LoadTestRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid JSON body")
	}
	if req.URL == "" {
		return badRequest(c, "url is required")
	}

	opts := loadtest.Options{
		URL:      req.URL,
		Method:   req.Method,
		Workers:  req.Workers,
		RPS:      req.RPS,
		Count:    req.Count,
		Duration: time.Duration(req.DurationMs) * time.Millisecond,
		Timeout:  msOr(req.TimeoutMs, 10*time.Second),
		Headers:  req.Headers,
		BypassCF: req.BypassCloudflare,
	}

	// Bind the run to the request context so a client disconnect stops it.
	stats, err := h.svc.LoadTest(c.Request().Context(), opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, stats)
}

func badRequest(c echo.Context, msg string) error {
	return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: msg})
}

func msOr(ms int, fallback time.Duration) time.Duration {
	if ms <= 0 {
		return fallback
	}
	return time.Duration(ms) * time.Millisecond
}
