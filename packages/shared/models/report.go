package models

type ReportData struct {
	GeneratedAt  string         `json:"generated_at"`
	ReportType   string         `json:"report_type"`
	Summary      Summary        `json:"summary"`
	Checks       []HealthResult `json:"checks,omitempty"`
	RequestStats RequestStats   `json:"request_stats,omitempty"`
}

type Summary struct {
	TotalChecks   int     `json:"total_checks"`
	Successful    int     `json:"successful"`
	Failed        int     `json:"failed"`
	UptimePercent float64 `json:"uptime_percentage"`
}
