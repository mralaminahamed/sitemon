# Portman

Portfolio HTTP Request Manager & Health Check Tool

## Overview

Portman is a Go-based CLI tool for HTTP request management and portfolio site health monitoring. It's designed for terminal usage with support for health checks, high-volume request testing, and report generation.

## Features

- **Health Checks** - Monitor URL availability and response times
- **Watch Mode** - Continuous health monitoring at specified intervals
- **High-Volume Request Testing** - Load testing with configurable workers and RPS limits
- **Detailed Latency Stats** - P50, P90, P95, P99 percentiles
- **Report Generation** - Export health check and load test results to JSON
- **Rate Limiting** - Built-in RPS limiter to protect target servers
- **Structured Logging** - JSON logging via zerolog with file output

## Installation

```bash
# Clone the repository
git clone https://github.com/mralaminahamed/portman.git
cd portman

# Install dependencies
go mod tidy

# Build the binary
go build -o portman .

# Or install globally
go install
```

## Quick Start

### Health Check

```bash
portman check --url https://example.com
```

### Watch Mode

```bash
portman check --url https://example.com --watch --interval 30s
```

### Load Testing

```bash
portman request --url https://example.com --workers 100 --rps 500 --count 10000
```

### Generate Report

```bash
portman report --output report.json
```

## Commands

| Command | Description |
|---------|-------------|
| `portman check` | Check health of URLs with status monitoring |
| `portman request` | Send high-volume HTTP requests for load testing |
| `portman report` | Generate JSON reports from check results or stats |
| `portman version` | Show version information |

For detailed usage and examples, see [DOCS.md](DOCS.md).

## Configuration

Configuration can be set via `config.yaml` or environment variables:

```yaml
url: ""
workers: 10
rps: 100
timeout: 10
log_level: info
log_file: ""
method: "GET"
tls_insecure: false
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `PORTMAN_URL` | Default URL |
| `PORTMAN_WORKERS` | Default workers |
| `PORTMAN_RPS` | Default RPS |
| `PORTMAN_TIMEOUT` | Default timeout |
| `PORTMAN_LOG_LEVEL` | Log level |
| `PORTMAN_LOG_FILE` | Log file path |
| `PORTMAN_TLS_INSECURE` | Skip TLS verification |

## Safe Testing Thresholds

| VPS RAM | Max Workers (No Cache) | Max Workers (With Cache) |
|---------|------------------------|--------------------------|
| 1 GB    | 10–15                  | 50–100                   |
| 2 GB    | 20–30                  | 100–200                  |
| 4 GB    | 40–60                  | 300–500                  |
| 8 GB    | 80–120                 | 600–1,000                |

## Tech Stack

- **CLI** - spf13/cobra
- **Config** - spf13/viper
- **HTTP Client** - go-resty/resty/v2
- **TUI** - charmbracelet/bubbletea
- **Logging** - rs/zerolog
- **Rate Limiting** - golang.org/x/time/rate
- **Concurrency** - golang.org/x/sync/errgroup

## Documentation

For complete documentation, see [DOCS.md](DOCS.md).

## License

MIT
