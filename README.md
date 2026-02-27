# Portman

Portfolio HTTP Request Manager & Health Check Tool

## Overview

Portman is a Go-based CLI tool for HTTP request management and portfolio site health monitoring. It provides health checks, continuous monitoring, high-volume request testing, SSL certificate monitoring, scheduled checks, and comprehensive reporting.

## Features

- **Health Checks** - Monitor URL availability and response times
- **Watch Mode** - Continuous health monitoring at specified intervals
- **SSL Certificate Monitoring** - Check expiry, issuer, protocol details
- **Scheduled Checks** - Cron-based automated health monitoring
- **Real-time Dashboard** - Interactive TUI for live monitoring
- **Webhook Alerts** - Slack, Discord, Telegram notifications
- **High-Volume Request Testing** - Load testing with configurable workers and RPS limits
- **Detailed Latency Stats** - P50, P90, P95, P99 percentiles
- **History Database** - SQLite storage for historical data
- **Content Validation** - Verify response contains specific text

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

### SSL Certificate Check

```bash
portman ssl --url https://example.com
```

### Watch Mode with Alerts

```bash
portman check --url https://example.com --watch --webhook https://hooks.slack.com/xxx
```

### Load Testing

```bash
portman request --url https://example.com --workers 100 --rps 500 --count 10000
```

### Dashboard

```bash
portman dashboard --url https://example.com
```

## Commands

| Command | Description |
|---------|-------------|
| `check` | Check health of URLs with optional watch mode |
| `ssl` | Check SSL certificate details |
| `schedule` | Run checks on cron schedule |
| `dashboard` | Interactive TUI dashboard |
| `history` | View check history from SQLite |
| `request` | High-volume load testing |
| `report` | Generate JSON reports |

## Configuration

Configuration can be set via `config.yaml`, `.env`, or environment variables.

### Environment Variables

| Variable | Description |
|----------|-------------|
| `PORTMAN_URL` | Default URL |
| `PORTMAN_WORKERS` | Default workers |
| `PORTMAN_RPS` | Default RPS |
| `PORTMAN_TIMEOUT` | Default timeout |
| `PORTMAN_LOG_LEVEL` | Log level |
| `PORTMAN_LOG_FILE` | Log file path |
| `PORTMAN_WEBHOOK_URL` | Webhook URL for alerts |
| `PORTMAN_TLS_INSECURE` | Skip TLS verification |

## Documentation

For complete documentation, see [DOCS.md](DOCS.md).

## License

MIT
