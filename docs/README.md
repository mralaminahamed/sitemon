# Sitemon Documentation

Welcome to the Sitemon documentation.

## Available Commands

| Command | Description | Use Case |
|---------|-------------|----------|
| [check](check.md) | Health check URL availability | Monitor website uptime |
| [ssl](ssl.md) | Check SSL certificate details | Monitor SSL expiry |
| [schedule](schedule.md) | Run checks on cron schedule | Automated monitoring |
| [dashboard](dashboard.md) | Interactive TUI dashboard | Real-time monitoring |
| [history](history.md) | View check history | Analyze past results |
| [request](request.md) | High-volume load testing | Stress test endpoints |
| [report](report.md) | Generate JSON reports | Export analysis |

## Quick Start

```bash
# Health check
sitemon check --url https://example.com

# Continuous monitoring
sitemon check --url https://example.com --watch --interval 30s

# Load testing
sitemon request --url https://example.com --workers 100 --rps 500

# SSL check
sitemon ssl --url https://example.com

# Dashboard
sitemon dashboard --url https://example.com
```

## Load Testing Examples

```bash
# 10K requests
sitemon request -u https://example.com -w 50 -r 500 -n 10000

# 100K requests  
sitemon request -u https://example.com -w 100 -r 1000 -n 100000

# 1 Million requests
sitemon request -u https://example.com -w 200 -r 2000 -n 1000000

# 1 Million requests with Cloudflare bypass
sitemon request -u https://codecept.io -w 100 -r 1000 -n 1000000 --bypass-cloudflare
```

## Guides

- [Configuration](configuration.md) - Configuration options

## Quick Links

- [Installation](../README.md)
- [GitHub Repository](https://github.com/mralaminahamed/sitemon)
