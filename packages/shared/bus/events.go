package bus

// Event and RPC payloads carried over the bus. Kept as plain structs with JSON
// tags — the wire format is JSON (see Publish/Request).

// CheckJob is emitted by the scheduler and consumed by checkers.
type CheckJob struct {
	URL              string `json:"url"`
	Method           string `json:"method,omitempty"`
	TimeoutMs        int    `json:"timeout_ms,omitempty"`
	BypassCloudflare bool   `json:"bypass_cloudflare,omitempty"`
}

// AlertEvent is published by the notifier when it decides a transition is
// alert-worthy (audit trail; gateway may surface it).
type AlertEvent struct {
	URL        string `json:"url"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Type       string `json:"type"` // "down" | "recovery"
	Timestamp  string `json:"timestamp"`
}

// RunCheckRequest / RunLoadTestRequest are the checker RPC inputs. The
// responses reuse the shared domain models (HealthResult, RequestStats).
type RunCheckRequest struct {
	URL              string `json:"url"`
	TimeoutMs        int    `json:"timeout_ms,omitempty"`
	BypassCloudflare bool   `json:"bypass_cloudflare,omitempty"`
}

type RunAnalyzeRequest struct {
	URL   string `json:"url"`
	Limit int    `json:"limit,omitempty"`
}

type RunLoadTestRequest struct {
	URL              string            `json:"url"`
	Method           string            `json:"method,omitempty"`
	Workers          int               `json:"workers,omitempty"`
	RPS              int               `json:"rps,omitempty"`
	Count            int               `json:"count,omitempty"`
	DurationMs       int               `json:"duration_ms,omitempty"`
	TimeoutMs        int               `json:"timeout_ms,omitempty"`
	Headers          map[string]string `json:"headers,omitempty"`
	BypassCloudflare bool              `json:"bypass_cloudflare,omitempty"`
}
