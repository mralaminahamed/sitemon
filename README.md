# Sitemon

Site Health Monitor & HTTP Request Manager

## Overview

Sitemon is a Go-based CLI tool for HTTP request management and website health monitoring. It works great for monitoring portfolios, personal sites, and any web services.

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
git clone https://github.com/mralaminahamed/sitemon.git
cd sitemon

# Install dependencies
go mod tidy

# Build the binary
go build -o sitemon .

# Or install globally
go install
```

## Quick Start

### Health Check

```bash
sitemon check --url https://example.com
```

### SSL Certificate Check

```bash
sitemon ssl --url https://example.com
```

### Watch Mode with Alerts

```bash
sitemon check --url https://example.com --watch --webhook https://hooks.slack.com/xxx
```

### Load Testing

```bash
sitemon request --url https://example.com --workers 100 --rps 500 --count 10000
```

### Dashboard

```bash
sitemon dashboard --url https://example.com
```

## Commands

| Command | Description |
|---------|-------------|
| [check](docs/check.md) | Check health of URLs with optional watch mode |
| [ssl](docs/ssl.md) | Check SSL certificate details |
| [schedule](docs/schedule.md) | Run checks on cron schedule |
| [dashboard](docs/dashboard.md) | Interactive TUI dashboard |
| [history](docs/history.md) | View check history from SQLite |
| [request](docs/request.md) | High-volume load testing |
| [report](docs/report.md) | Generate JSON reports |

## Documentation

For detailed documentation, see the [docs](docs/) directory:

- [Check Command](docs/check.md)
- [SSL Command](docs/ssl.md)
- [Schedule Command](docs/schedule.md)
- [Dashboard Command](docs/dashboard.md)
- [History Command](docs/history.md)
- [Request Command](docs/request.md)
- [Report Command](docs/report.md)
- [Configuration](docs/configuration.md)

## Configuration

Configuration can be set via `config.yaml`, `.env`, or environment variables.

See [Configuration](docs/configuration.md) for details.

## License

MIT
