// Package handler holds the Echo HTTP handlers. Handlers validate input, call
// the service layer, and shape the JSON response — no business logic here.
package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/dto"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/readmodel"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/service"
	"github.com/mralaminahamed/sitemon/packages/shared/loadtest"
	"github.com/mralaminahamed/sitemon/packages/shared/urlguard"
)

// Load-test hard ceilings — protect the platform from being used to hammer a
// target with unbounded workers/rps/count/duration.
const (
	maxWorkers  = 500
	maxRPS      = 5000
	maxCount    = 1_000_000
	maxDuration = 10 * time.Minute
)

type Handler struct {
	svc *service.Service
	rm  *readmodel.ReadModel
}

func New(svc *service.Service, rm *readmodel.ReadModel) *Handler {
	return &Handler{svc: svc, rm: rm}
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
			"GET  /api/history?url=&limit=",
			"GET  /api/stats?url=",
			"GET  /api/alerts?url=&limit=",
			"GET  /api/analyze?url=",
			"GET  /api/monitors",
			"POST /api/monitors",
			"DELETE /api/monitors?url=",
			"POST /api/checks",
			"GET  /api/ssl?url=",
			"POST /api/loadtest",
		},
	})
}

// Status returns the latest health result per URL.
func (h *Handler) Status(c echo.Context) error {
	results, err := h.rm.Status(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"results": dto.FromHealthResults(results)})
}

// History returns stored check history, optionally filtered by ?url=.
func (h *Handler) History(c echo.Context) error {
	limit, _ := strconv.ParseInt(c.QueryParam("limit"), 10, 64)
	results, err := h.rm.History(c.Request().Context(), c.QueryParam("url"), limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"results": dto.FromHealthResults(results)})
}

// Monitors lists tracked URLs.
func (h *Handler) Monitors(c echo.Context) error {
	monitors, err := h.rm.ListMonitors(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"monitors": monitors})
}

// AddMonitor adds a tracked URL.
func (h *Handler) AddMonitor(c echo.Context) error {
	var req struct {
		URL string `json:"url"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid JSON body")
	}
	url := normalizeURL(req.URL)
	if url == "" {
		return badRequest(c, "url is required")
	}
	if err := urlguard.Check(url); err != nil {
		return badRequest(c, err.Error())
	}
	m, err := h.rm.AddMonitor(c.Request().Context(), url)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{Error: err.Error()})
	}
	h.kickCheck(url)
	return c.JSON(http.StatusCreated, m)
}

// DeleteMonitor removes a tracked URL (?url=).
func (h *Handler) DeleteMonitor(c echo.Context) error {
	url := c.QueryParam("url")
	if url == "" {
		return badRequest(c, "url query parameter is required")
	}
	if err := h.rm.DeleteMonitor(c.Request().Context(), url); err != nil {
		return c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{Error: err.Error()})
	}
	return c.NoContent(http.StatusNoContent)
}

// Alerts returns recent alert history, optionally filtered by ?url=.
func (h *Handler) Alerts(c echo.Context) error {
	limit, _ := strconv.ParseInt(c.QueryParam("limit"), 10, 64)
	alerts, err := h.rm.Alerts(c.Request().Context(), c.QueryParam("url"), limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"alerts": alerts})
}

// Analyze returns an AI incident analysis for ?url=.
func (h *Handler) Analyze(c echo.Context) error {
	url := c.QueryParam("url")
	if url == "" {
		return badRequest(c, "url query parameter is required")
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	out, err := h.svc.Analyze(url, limit)
	if err != nil {
		if err == service.ErrNoBus {
			return c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusBadGateway, dto.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, out)
}

// Stats returns aggregate uptime/latency for ?url=.
func (h *Handler) Stats(c echo.Context) error {
	st, err := h.rm.Stats(c.Request().Context(), c.QueryParam("url"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, st)
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
	for _, u := range req.URLs {
		if err := urlguard.Check(u); err != nil {
			return badRequest(c, err.Error())
		}
	}

	timeout := msOr(req.TimeoutMs, 10*time.Second)
	results := h.svc.CheckURLs(req.URLs, timeout, req.BypassCloudflare)
	return c.JSON(http.StatusOK, echo.Map{"results": dto.FromHealthResults(results)})
}

// SSL returns certificate details for ?url=.
func (h *Handler) SSL(c echo.Context) error {
	url := c.QueryParam("url")
	if url == "" {
		return badRequest(c, "url query parameter is required")
	}
	if err := urlguard.Check(url); err != nil {
		return badRequest(c, err.Error())
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
	if err := urlguard.Check(req.URL); err != nil {
		return badRequest(c, err.Error())
	}

	opts := loadtest.Options{
		URL:      req.URL,
		Method:   req.Method,
		Workers:  clamp(req.Workers, maxWorkers),
		RPS:      clamp(req.RPS, maxRPS),
		Count:    clamp(req.Count, maxCount),
		Duration: clampDur(time.Duration(req.DurationMs)*time.Millisecond, maxDuration),
		Timeout:  msOr(req.TimeoutMs, 10*time.Second),
		Headers:  req.Headers,
		BypassCF: req.BypassCloudflare,
	}

	// Bind the run to the request context so a client disconnect stops it.
	stats, err := h.svc.LoadTest(c.Request().Context(), opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, dto.FromRequestStats(stats))
}

// kickCheck runs a first check for a new monitor in the background so it shows up
// immediately. Prefers the bus (normal pipeline → live broadcast); falls back to
// an in-process check when there is no bus.
func (h *Handler) kickCheck(url string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := h.svc.KickCheck(ctx, url); err != nil {
			for _, r := range h.svc.CheckURLs([]string{url}, 10*time.Second, false) {
				h.rm.PutResult(ctx, r)
			}
		}
	}()
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

func normalizeURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return u
	}
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		return "https://" + u
	}
	return u
}

func clamp(v, max int) int {
	if v > max {
		return max
	}
	return v
}

func clampDur(v, max time.Duration) time.Duration {
	if v > max {
		return max
	}
	return v
}
