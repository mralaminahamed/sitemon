// Package analyzer derives uptime/latency metrics and anomaly flags from check
// history. Statistical only — the LLM layer turns this into prose.
package analyzer

import (
	"sort"

	"github.com/mralaminahamed/sitemon/packages/shared/models"
)

type Analysis struct {
	URL       string   `json:"url"`
	Window    int      `json:"window"`
	UptimePct float64  `json:"uptime_pct"`
	ErrorRate float64  `json:"error_rate"`
	AvgMs     float64  `json:"avg_ms"`
	P95Ms     float64  `json:"p95_ms"`
	Current   string   `json:"current_status"`
	Anomalies []string `json:"anomalies"`
	Severity  string   `json:"severity"` // ok | warn | critical
	Summary   string   `json:"summary,omitempty"`
}

func Analyze(url string, results []models.HealthResult) Analysis {
	a := Analysis{URL: url, Window: len(results), Anomalies: []string{}, Severity: "ok", Current: "UNKNOWN"}
	if len(results) == 0 {
		return a
	}

	a.Current = results[0].Status // Mongo returns newest first

	var up, bad int
	ms := make([]float64, 0, len(results))
	for _, r := range results {
		if r.Status == "UP" {
			up++
		}
		if r.Status == "DOWN" || r.Error != "" {
			bad++
		}
		ms = append(ms, float64(r.ResponseTime)/1e6)
	}
	n := float64(len(results))
	a.UptimePct = float64(up) / n * 100
	a.ErrorRate = float64(bad) / n

	sorted := append([]float64(nil), ms...)
	sort.Float64s(sorted)
	var sum float64
	for _, v := range sorted {
		sum += v
	}
	a.AvgMs = sum / n
	a.P95Ms = percentile(sorted, 0.95)

	latest := ms[0]
	if a.AvgMs > 0 && latest > a.AvgMs*2 {
		a.Anomalies = append(a.Anomalies, "latency spike: latest above 2x average")
	}
	if a.ErrorRate > 0.2 {
		a.Anomalies = append(a.Anomalies, "elevated error rate")
	}
	if a.UptimePct < 100 && a.UptimePct >= 90 {
		a.Anomalies = append(a.Anomalies, "intermittent downtime")
	}

	switch {
	case a.Current == "DOWN" || a.UptimePct < 90:
		a.Severity = "critical"
	case a.Current == "WARNING" || a.ErrorRate > 0.1 || len(a.Anomalies) > 0:
		a.Severity = "warn"
	}
	return a
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	i := int(float64(len(sorted)) * p)
	if i >= len(sorted) {
		i = len(sorted) - 1
	}
	return sorted[i]
}
