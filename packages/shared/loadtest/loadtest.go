// Package loadtest is a small, reusable HTTP load generator. It is shared by
// the CLI's `request` command and the gateway's /loadtest endpoint (and, later,
// the checker service). The engine carries the correctness fixes made in
// Phase 0: a single shared rate limiter, mutex-guarded min/max, and prompt
// cancellation via context.
package loadtest

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	shttp "github.com/mralaminahamed/sitemon/packages/shared/http"
	"github.com/mralaminahamed/sitemon/packages/shared/models"
	"github.com/mralaminahamed/sitemon/packages/shared/monitor"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

// Hard ceilings shared by every entrypoint (the gateway edge and the checker
// bus path) so the platform can't be turned into an unbounded load generator.
const (
	MaxWorkers  = 500
	MaxRPS      = 5000
	MaxCount    = 1_000_000
	MaxDuration = 10 * time.Minute
)

// Clamp caps the run parameters to the platform ceilings. Called at every
// entrypoint so bus callers can't bypass the gateway's edge limits.
func (o *Options) Clamp() {
	if o.Workers > MaxWorkers {
		o.Workers = MaxWorkers
	}
	if o.RPS > MaxRPS {
		o.RPS = MaxRPS
	}
	if o.Count > MaxCount {
		o.Count = MaxCount
	}
	if o.Duration > MaxDuration {
		o.Duration = MaxDuration
	}
}

// Options configures a load run. Either Count or Duration bounds the run;
// when Duration > 0 it wins and the run is time-boxed.
type Options struct {
	URL      string
	Method   string
	Workers  int
	RPS      int
	Count    int
	Duration time.Duration
	Timeout  time.Duration
	Headers  map[string]string
	BypassCF bool
}

func (o *Options) applyDefaults() {
	if o.Method == "" {
		o.Method = "GET"
	}
	if o.Workers <= 0 {
		o.Workers = 10
	}
	if o.RPS <= 0 {
		o.RPS = 100
	}
	if o.Count <= 0 && o.Duration <= 0 {
		o.Count = 100
	}
	if o.Timeout <= 0 {
		o.Timeout = 10 * time.Second
	}
}

// Run executes the load test and returns aggregate statistics. It respects
// ctx cancellation (the caller can wire it to a signal or an HTTP request
// context to stop early).
func Run(ctx context.Context, opts Options) (models.RequestStats, error) {
	opts.applyDefaults()

	clientOpts := []shttp.ClientOption{shttp.WithTimeout(opts.Timeout)}
	if opts.BypassCF {
		clientOpts = append(clientOpts, shttp.WithCloudflareBypass())
	}
	for k, v := range opts.Headers {
		clientOpts = append(clientOpts, shttp.WithHeader(k, v))
	}
	client := shttp.NewFastClient(opts.Timeout, clientOpts...)

	var (
		successCount, failureCount int64
		totalLatency, totalSize    int64
		minSize                    int64 = -1
		maxSize                    int64
		sizeMu                     sync.Mutex
		latencies                  []time.Duration
		latMu                      sync.Mutex
		statusCodes                = map[int]int{}
		codeMu                     sync.Mutex
		sentCount                  int64
	)

	g := errgroup.Group{}
	g.SetLimit(opts.Workers)

	limiter := rate.NewLimiter(rate.Limit(opts.RPS), opts.RPS)
	start := time.Now()
	durationMode := opts.Duration > 0

	send := func() shttp.RequestResult {
		switch opts.Method {
		case "HEAD":
			return client.SendHead(opts.URL)
		case "POST":
			return client.SendPost(opts.URL, nil)
		case "PUT":
			return client.SendPut(opts.URL, nil)
		case "PATCH":
			return client.SendPatch(opts.URL, nil)
		case "DELETE":
			return client.SendDelete(opts.URL)
		case "OPTIONS":
			return client.SendOptions(opts.URL)
		default:
			return client.SendGet(opts.URL)
		}
	}

	for i := 0; ; i++ {
		if !durationMode && i >= opts.Count {
			break
		}
		if ctx.Err() != nil {
			break
		}
		if durationMode && time.Since(start) >= opts.Duration {
			break
		}
		if err := limiter.Wait(ctx); err != nil {
			break
		}

		g.Go(func() error {
			r := send()

			atomic.AddInt64(&sentCount, 1)
			atomic.AddInt64(&totalLatency, r.ResponseTime.Milliseconds())
			atomic.AddInt64(&totalSize, r.ResponseSize)

			sizeMu.Lock()
			if minSize == -1 || r.ResponseSize < minSize {
				minSize = r.ResponseSize
			}
			if r.ResponseSize > maxSize {
				maxSize = r.ResponseSize
			}
			sizeMu.Unlock()

			latMu.Lock()
			latencies = append(latencies, r.ResponseTime)
			latMu.Unlock()

			codeMu.Lock()
			statusCodes[r.StatusCode]++
			codeMu.Unlock()

			if r.Error != nil || r.StatusCode >= 400 {
				atomic.AddInt64(&failureCount, 1)
			} else {
				atomic.AddInt64(&successCount, 1)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil && err != context.Canceled {
		return models.RequestStats{}, err
	}

	elapsed := time.Since(start)
	total := atomic.LoadInt64(&sentCount)
	if total == 0 {
		return models.RequestStats{}, context.Canceled
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	lat := monitor.CalculateLatencyStats(latencies)

	if minSize == -1 {
		minSize = 0
	}

	return models.RequestStats{
		TotalRequests:   int(total),
		Successful:      int(successCount),
		Failed:          int(failureCount),
		Duration:        elapsed,
		RequestsPerSec:  float64(total) / elapsed.Seconds(),
		SuccessRate:     float64(successCount) / float64(total) * 100,
		AvgLatency:      lat.Avg,
		MinLatency:      lat.Min,
		MaxLatency:      lat.Max,
		P50Latency:      lat.P50,
		P90Latency:      lat.P90,
		P95Latency:      lat.P95,
		P99Latency:      lat.P99,
		TargetURL:       opts.URL,
		Method:          opts.Method,
		Workers:         opts.Workers,
		RPS:             opts.RPS,
		ResponseSizeAvg: atomic.LoadInt64(&totalSize) / total,
		ResponseSizeMin: minSize,
		ResponseSizeMax: maxSize,
		StatusCodes:     statusCodes,
		DurationMode:    durationMode,
	}, nil
}
