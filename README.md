# Portman

Portfolio HTTP Request Manager & Health Check Tool

## Overview

Portman is a Go-based CLI tool for HTTP request management and portfolio site health monitoring. It's designed for terminal usage with support for health checks, high-volume request testing, and report generation.

## Features

- **Health Checks** - Monitor URL availability and response times
- **Watch Mode** - Continuous health monitoring at specified intervals
- **High-Volume Request Testing** - Load testing with configurable workers and RPS limits
- **Report Generation** - Export health check results to JSON
- **Rate Limiting** - Built-in RPS limiter to protect target servers
- **Structured Logging** - JSON logging via zerolog

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

## Usage

### Health Check

Check the health of a single URL:

```bash
portman check --url https://example.com
```

### Watch Mode

Monitor a URL continuously:

```bash
portman check --url https://example.com --watch --interval 30s
```

### High-Volume Request Testing

Send bulk HTTP requests with rate limiting:

```bash
portman request --url https://example.com --workers 100 --rps 500 --count 10000
```

Options:
- `--url, -u` - Target URL (required)
- `--workers, -w` - Number of concurrent workers (default: 10)
- `--rps, -r` - Requests per second limit (default: 100)
- `--count, -n` - Total number of requests (default: 1000)
- `--timeout` - Request timeout (default: 10s)

### Generate Report

Export health check results:

```bash
portman report --output report.json
```

## Configuration

Configuration can be set via `config.yaml` or environment variables:

```yaml
url: ""
workers: 10
rps: 100
timeout: 10
log_level: info
```

### CLI Flags

- `--config, -c` - Config file path
- `--verbose, -v` - Enable verbose output
- `--log-level` - Log level (debug, info, warn, error)

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

## License

MIT
