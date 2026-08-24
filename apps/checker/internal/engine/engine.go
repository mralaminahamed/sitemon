// Package engine is the checker's core: it runs a single health check or a
// load test by composing the shared engines. Kept separate from transport
// (NATS) so it stays trivially testable.
package engine

import (
	"context"
	"strings"
	"time"

	shttp "github.com/mralaminahamed/sitemon/packages/shared/http"
	"github.com/mralaminahamed/sitemon/packages/shared/loadtest"
	"github.com/mralaminahamed/sitemon/packages/shared/models"
	"github.com/mralaminahamed/sitemon/packages/shared/monitor"
)

// Check runs one health check and returns the result.
func Check(url string, timeout time.Duration, bypassCF bool) models.HealthResult {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	url = normalize(url)

	opts := []shttp.ClientOption{shttp.WithTimeout(timeout)}
	if bypassCF {
		opts = append(opts, shttp.WithCloudflareBypass())
	}
	client := shttp.NewClient(timeout, opts...)

	// HealthChecker.Check never returns a nil result or a non-nil error.
	res, _ := monitor.NewHealthChecker(client).Check(url)
	return *res
}

// LoadTest runs a load test.
func LoadTest(ctx context.Context, opts loadtest.Options) (models.RequestStats, error) {
	opts.URL = normalize(opts.URL)
	return loadtest.Run(ctx, opts)
}

func normalize(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return u
	}
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		return "https://" + u
	}
	return u
}
