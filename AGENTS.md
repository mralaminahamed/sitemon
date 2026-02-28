# Sitemon - AGENTS.md

## Project Overview

**Sitemon** (Site Monitor) is a Go-based CLI tool for HTTP request management and website health monitoring. It's designed for terminal usage with support for health checks, high-volume request testing, and report generation.

- **Language**: Go 1.22+
- **Type**: Terminal CLI Application
- **Module**: `github.com/mralaminahamed/sitemon`

## Tech Stack

| Category | Package | Version |
|----------|---------|---------|
| CLI | `github.com/spf13/cobra` | v1.8.0 |
| Config | `github.com/spf13/viper` | v1.18.0 |
| HTTP Client | `github.com/go-resty/resty/v2` | v2.12.0 |
| TUI | `github.com/charmbracelet/bubbletea` | v0.26.0 |
| Styling | `github.com/charmbracelet/lipgloss` | v0.10.0 |
| Tables | `github.com/olekukonko/tablewriter` | v0.5.6 |
| Progress | `github.com/schollz/progressbar/v3` | v3.14.2 |
| Logging | `github.com/rs/zerolog` | v1.32.0 |
| Rate Limit | `golang.org/x/time/rate` | v0.5.0 |
| Concurrency | `golang.org/x/sync/errgroup` | v0.6.0 |

## Project Structure

```
sitemon/
├── cmd/
│   ├── root.go          # Root cobra command, global flags
│   ├── check.go         # Health check subcommand
│   ├── request.go       # HTTP request management subcommand
│   └── report.go        # Report generation subcommand
├── internal/
│   ├── http/
│   │   └── client.go    # Resty HTTP client wrapper
│   ├── config/
│   │   └── config.go    # Viper config loader
│   ├── monitor/
│   │   └── health.go    # Health check logic
│   └── logger/
│       └── logger.go    # Zerolog setup
├── config.yaml          # Default configuration file
├── go.mod
├── go.sum
└── main.go
```

## CLI Commands

```bash
# Health check
sitemon check --url https://yourdomain.com

# Watch mode (continuous health monitoring)
sitemon check --watch --interval 30s

# High-volume request testing
sitemon request --url https://yourdomain.com --workers 100 --rps 500

# Generate report
sitemon report --output report.json
```

## Key Design Principles

1. **Concurrency-first** — Goroutines + errgroup for scalable request handling
2. **Rate-limit-aware** — Built-in RPS limiter to protect target VPS
3. **Config-driven** — All parameters via config.yaml or CLI flags
4. **Structured logging** — JSON logs via zerolog for parsing and reporting
5. **Single binary** — Compiled Go binary, no runtime dependencies
6. **Cross-platform** — Runs on Linux, macOS, Windows

## Dependency Responsibilities

| Dependency | Responsibility |
|------------|----------------|
| `cobra` | Parses CLI commands, flags, and subcommands |
| `viper` | Loads config from config.yaml, ENV variables, or flags |
| `resty` | Executes HTTP requests with timeout, retry, and header support |
| `bubbletea` | Renders live TUI dashboard for real-time monitoring |
| `lipgloss` | Styles TUI components (colors, borders, layout) |
| `tablewriter` | Formats response data as structured tables in terminal |
| `progressbar` | Displays progress during bulk/high-volume request runs |
| `zerolog` | Logs structured JSON output to file and stdout |
| `x/time/rate` | Token bucket rate limiter to control RPS |
| `x/sync/errgroup` | Manages goroutine lifecycle and aggregates errors |

## Development Guidelines

- Run `go mod tidy` after adding dependencies
- Use structured logging via zerolog (JSON format)
- All CLI parameters should be configurable via config.yaml
- Implement proper error handling with errgroup for concurrent operations
- Add rate limiting to prevent accidental self-DDoS during testing

## Testing Safe Thresholds

| VPS RAM | Max Safe Workers (No Cache) | Max Safe Workers (With Cache) |
|---------|----------------------------|-------------------------------|
| 1 GB | 10–15 | 50–100 |
| 2 GB | 20–30 | 100–200 |
| 4 GB | 40–60 | 300–500 |
| 8 GB | 80–120 | 600–1,000 |
