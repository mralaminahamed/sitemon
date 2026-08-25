// Package service wraps the check/ssl/load engines so handlers stay thin.
//
// When a bus is present, health checks and load tests are dispatched to the
// checker service via NATS request-reply (the Phase 2 split). When it is nil,
// the engines run in-process — so the gateway still works standalone, exactly
// as in Phase 1. SSL stays in-process either way.
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mralaminahamed/sitemon/packages/shared/bus"
	shttp "github.com/mralaminahamed/sitemon/packages/shared/http"
	"github.com/mralaminahamed/sitemon/packages/shared/loadtest"
	"github.com/mralaminahamed/sitemon/packages/shared/models"
	"github.com/mralaminahamed/sitemon/packages/shared/monitor"
	"github.com/mralaminahamed/sitemon/packages/shared/ssl"
)

// ErrNoBus is returned by bus-only operations when running in-process.
var ErrNoBus = errors.New("feature requires the event bus (distributed mode)")

type Service struct {
	bus *bus.Bus // nil => in-process mode
}

func New(b *bus.Bus) *Service { return &Service{bus: b} }

// Distributed reports whether checks are dispatched to the checker service.
func (s *Service) Distributed() bool { return s.bus != nil }

// CheckURLs runs a health check for each URL and returns results in order.
func (s *Service) CheckURLs(urls []string, timeout time.Duration, bypassCF bool) []models.HealthResult {
	urls = normalize(urls)

	if s.bus == nil {
		return s.checkInProcess(urls, timeout, bypassCF)
	}

	results := make([]models.HealthResult, 0, len(urls))
	for _, u := range urls {
		var res models.HealthResult
		req := bus.RunCheckRequest{URL: u, TimeoutMs: int(timeout / time.Millisecond), BypassCloudflare: bypassCF}
		if err := s.bus.Request(bus.SubjectCheckerRun, req, &res, timeout+5*time.Second); err != nil {
			res = models.HealthResult{URL: u, Status: "DOWN", Error: err.Error()}
		}
		results = append(results, res)
	}
	return results
}

func (s *Service) checkInProcess(urls []string, timeout time.Duration, bypassCF bool) []models.HealthResult {
	opts := []shttp.ClientOption{shttp.WithTimeout(timeout)}
	if bypassCF {
		opts = append(opts, shttp.WithCloudflareBypass())
	}
	client := shttp.NewClient(timeout, opts...)
	return monitor.NewHealthChecker(client).CheckMultiple(urls)
}

// KickCheck publishes an immediate check.job so a newly added monitor is checked
// without waiting for the next scheduler tick. Requires the bus; returns ErrNoBus
// in-process (caller should fall back to a direct check).
func (s *Service) KickCheck(ctx context.Context, url string) error {
	if s.bus == nil {
		return ErrNoBus
	}
	return s.bus.Publish(ctx, bus.SubjectCheckJob, bus.CheckJob{URL: normalizeOne(url)})
}

// Analyze asks the ai service for an incident analysis of url. Requires the
// bus (distributed mode); returns an error otherwise.
func (s *Service) Analyze(url string, limit int) (map[string]any, error) {
	if s.bus == nil {
		return nil, ErrNoBus
	}
	var out map[string]any
	req := bus.RunAnalyzeRequest{URL: normalizeOne(url), Limit: limit}
	err := s.bus.Request(bus.SubjectAIAnalyze, req, &out, 35*time.Second)
	return out, err
}

// SSL returns certificate details for a single URL (always in-process).
func (s *Service) SSL(url string, timeout time.Duration) (*ssl.CertificateInfo, error) {
	return ssl.CheckCertificate(normalizeOne(url), timeout)
}

// LoadTest runs a load test, via the checker when distributed.
func (s *Service) LoadTest(ctx context.Context, opts loadtest.Options) (models.RequestStats, error) {
	opts.URL = normalizeOne(opts.URL)

	if s.bus == nil {
		return loadtest.Run(ctx, opts)
	}

	req := bus.RunLoadTestRequest{
		URL: opts.URL, Method: opts.Method, Workers: opts.Workers, RPS: opts.RPS,
		Count: opts.Count, DurationMs: int(opts.Duration / time.Millisecond),
		TimeoutMs: int(opts.Timeout / time.Millisecond), Headers: opts.Headers,
		BypassCloudflare: opts.BypassCF,
	}
	var stats models.RequestStats
	err := s.bus.Request(bus.SubjectCheckerLoadTest, req, &stats, loadTestRPCTimeout(opts))
	return stats, err
}

// loadTestRPCTimeout allows enough time for the whole run plus slack.
func loadTestRPCTimeout(o loadtest.Options) time.Duration {
	budget := 30 * time.Second
	if o.Duration > 0 {
		budget = o.Duration + 15*time.Second
	}
	return budget
}

func normalize(urls []string) []string {
	out := make([]string, len(urls))
	for i, u := range urls {
		out[i] = normalizeOne(u)
	}
	return out
}

func normalizeOne(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return u
	}
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		return "https://" + u
	}
	return u
}
