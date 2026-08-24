package loadtest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRunCountMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	stats, err := Run(context.Background(), Options{
		URL:     srv.URL,
		Method:  "GET",
		Workers: 4,
		RPS:     200,
		Count:   30,
		Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if stats.TotalRequests != 30 {
		t.Errorf("TotalRequests = %d, want 30", stats.TotalRequests)
	}
	if stats.Successful != 30 {
		t.Errorf("Successful = %d, want 30", stats.Successful)
	}
	if stats.StatusCodes[200] != 30 {
		t.Errorf("status 200 count = %d, want 30", stats.StatusCodes[200])
	}
	if !(stats.P50Latency <= stats.P99Latency) {
		t.Errorf("p50 %v > p99 %v", stats.P50Latency, stats.P99Latency)
	}
}

func TestRunCountsFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	stats, err := Run(context.Background(), Options{
		URL: srv.URL, Workers: 2, RPS: 100, Count: 10, Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if stats.Failed != 10 {
		t.Errorf("Failed = %d, want 10 (500s count as failures)", stats.Failed)
	}
}
