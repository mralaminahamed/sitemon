// Package dto holds the gateway's request/response shapes. Keeping them
// separate from the shared domain models lets the HTTP contract evolve
// independently of internal types.
package dto

// CheckRequest runs health checks against one or more URLs.
type CheckRequest struct {
	URLs             []string `json:"urls"`
	TimeoutMs        int      `json:"timeout_ms"`
	BypassCloudflare bool     `json:"bypass_cloudflare"`
}

// LoadTestRequest configures a load run.
type LoadTestRequest struct {
	URL              string            `json:"url"`
	Method           string            `json:"method"`
	Workers          int               `json:"workers"`
	RPS              int               `json:"rps"`
	Count            int               `json:"count"`
	DurationMs       int               `json:"duration_ms"`
	TimeoutMs        int               `json:"timeout_ms"`
	Headers          map[string]string `json:"headers"`
	BypassCloudflare bool              `json:"bypass_cloudflare"`
}

// ErrorResponse is the uniform error body.
type ErrorResponse struct {
	Error string `json:"error"`
}
