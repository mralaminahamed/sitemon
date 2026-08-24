package analyzer

import (
	"testing"
	"time"

	"github.com/mralaminahamed/sitemon/packages/shared/models"
)

func hr(status string, ms int, errStr string) models.HealthResult {
	return models.HealthResult{
		URL:          "https://example.com",
		Status:       status,
		ResponseTime: time.Duration(ms) * time.Millisecond,
		Error:        errStr,
	}
}

func TestAnalyzeAllUp(t *testing.T) {
	in := []models.HealthResult{hr("UP", 40, ""), hr("UP", 50, ""), hr("UP", 45, "")}
	a := Analyze("https://example.com", in)
	if a.Current != "UP" || a.Severity != "ok" {
		t.Errorf("current=%s severity=%s, want UP/ok", a.Current, a.Severity)
	}
	if a.UptimePct != 100 {
		t.Errorf("uptime=%v, want 100", a.UptimePct)
	}
}

func TestAnalyzeDownIsCritical(t *testing.T) {
	in := []models.HealthResult{hr("DOWN", 0, "dial error"), hr("UP", 50, ""), hr("UP", 55, "")}
	a := Analyze("https://example.com", in)
	if a.Current != "DOWN" || a.Severity != "critical" {
		t.Errorf("current=%s severity=%s, want DOWN/critical", a.Current, a.Severity)
	}
	if a.ErrorRate == 0 {
		t.Error("expected non-zero error rate")
	}
}

func TestAnalyzeLatencySpike(t *testing.T) {
	in := []models.HealthResult{hr("UP", 500, ""), hr("UP", 40, ""), hr("UP", 45, ""), hr("UP", 42, "")}
	a := Analyze("https://example.com", in)
	found := false
	for _, an := range a.Anomalies {
		if an == "latency spike: latest above 2x average" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected latency spike anomaly, got %v", a.Anomalies)
	}
}

func TestAnalyzeEmpty(t *testing.T) {
	a := Analyze("https://example.com", nil)
	if a.Window != 0 || a.Severity != "ok" {
		t.Errorf("empty analysis wrong: %+v", a)
	}
}
