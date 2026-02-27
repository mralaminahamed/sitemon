# Portman Documentation

## Overview

Portman is a CLI tool for HTTP request management and portfolio site health monitoring. It provides health checks, continuous monitoring, high-volume request testing, SSL certificate monitoring, scheduled checks, and comprehensive reporting.

## Installation

### From Source

```bash
git clone https://github.com/mralaminahamed/portman.git
cd portman
go install
```

### Build Binary

```bash
go build -o portman .
```

## Quick Start

### Check a URL

```bash
portman check --url https://example.com
```

Output:
```
[12:59:56] ✓ https://example.com - 200 (UP) in 44.055417ms

--- Summary ---
Total Checks: 1
Successful: 1
Failed: 0
Uptime: 100.00%
Avg Response Time: 44ms
Min Response Time: 44.055417ms
Max Response Time: 44.055417ms
```

---

## Commands

### Global Flags

These flags are available for all commands:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | config.yaml | Config file path |
| `--log-file` | - | - | Log file path |
| `--log-level` | - | info | Log level (debug, info, warn, error) |
| `--verbose` | `-v` | false | Enable verbose output |
| `--version` | - | false | Show version information |

### Version

Show version information:

```bash
portman version
```

---

### Health Check (`check`)

Check the health status of one or more URLs.

```bash
portman check [url...]
```

#### Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--url` | `-u` | [] | URL(s) to check (can be specified multiple times) |
| `--watch` | `-w` | false | Watch mode for continuous monitoring |
| `--interval` | - | 30s | Interval for watch mode |
| `--timeout` | - | 10s | Request timeout |
| `--json` | `-j` | false | Output in JSON format |
| `--output` | `-o` | - | Output file path |
| `--webhook` | - | - | Webhook URL for alerts (Slack, Discord, Telegram) |
| `--max-latency` | - | - | Maximum allowed latency before alert |
| `--contains` | - | - | Response must contain these strings |
| `--not-contains` | - | - | Response must NOT contain these strings |
| `--check-ssl` | - | false | Check SSL certificate |
| `--save` | - | false | Save results to history database |
| `--db` | - | - | Path to history database |

#### Examples

**Basic check:**
```bash
portman check -u https://example.com
```

**Check multiple URLs:**
```bash
portman check -u https://example.com -u https://google.com
portman check https://example.com https://google.com
```

**Watch mode with alerts:**
```bash
portman check -u https://example.com --watch --webhook https://hooks.slack.com/services/xxx
```

**Content validation:**
```bash
portman check -u https://example.com --contains "Welcome" --not-contains "Error"
```

**Check SSL certificate:**
```bash
portman check -u https://example.com --check-ssl
```

**Save to history:**
```bash
portman check -u https://example.com --save
```

---

### SSL Certificate (`ssl`)

Check SSL certificate details for URLs.

```bash
portman ssl [url...]
```

#### Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--url` | `-u` | [] | URL(s) to check |
| `--timeout` | - | 10s | Request timeout |

#### Examples

```bash
portman ssl -u https://example.com
portman ssl https://example.com https://google.com
```

#### Sample Output

```
=== SSL Certificate: https://example.com ===
Status:     ✓ Valid
Issuer:     SSL Corporation
Subject:    example.com
Valid From: 2026-02-13 18:53:48
Valid Until: 2026-05-14 18:57:50
Days Left:  76 days
Protocol:   TLS 772
Cipher:     0x1301
```

---

### Schedule (`schedule`)

Run health checks on a cron schedule.

```bash
portman schedule --cron "*/5 * * * *" --url https://example.com
```

#### Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--cron` | - | - | Cron expression (e.g., `*/5 * * * *`) |
| `--url` | `-u` | [] | URL(s) to check |
| `--timeout` | - | 10s | Request timeout |
| `--webhook` | - | - | Webhook URL for alerts |

#### Cron Examples

```bash
# Every 5 minutes
portman schedule --cron "*/5 * * * *" -u https://example.com

# Every hour
portman schedule --cron "0 * * * *" -u https://example.com

# Every day at midnight
portman schedule --cron "0 0 * * *" -u https://example.com

# With webhook notifications
portman schedule --cron "*/15 * * * *" -u https://example.com --webhook https://hooks.slack.com/xxx
```

---

### Dashboard (`dashboard`)

Run interactive TUI dashboard for real-time monitoring.

```bash
portman dashboard --url https://example.com
```

#### Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--url` | `-u` | [] | URL(s) to monitor |
| `--interval` | - | 5s | Refresh interval |
| `--timeout` | - | 10s | Request timeout |

#### Examples

```bash
portman dashboard -u https://example.com
portman dashboard -u https://example.com -u https://google.com --interval 10s
```

---

### History (`history`)

View health check history from SQLite database.

```bash
portman history
```

#### Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--db` | - | ~/.portman/history.db | Path to history database |
| `--limit` | `-n` | 50 | Number of records to show |
| `--from` | - | - | Start date (YYYY-MM-DD) |
| `--to` | - | - | End date (YYYY-MM-DD) |
| `--stats` | - | - | Show statistics for URL |
| `--export` | - | false | Export all history as JSON |
| `--output` | `-o` | - | Output file for export |

#### Examples

```bash
# View recent history
portman history

# Show last 100 records
portman history --limit 100

# Show statistics for URL
portman history --stats https://example.com

# Export to JSON
portman history --export -o history.json

# Filter by date
portman history --from 2026-01-01 --to 2026-02-27
```

---

### Request (`request`)

Send high-volume HTTP requests for load testing.

```bash
portman request [url]
```

#### Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--url` | `-u` | - | Target URL (required) |
| `--workers` | `-w` | 10 | Number of concurrent workers |
| `--rps` | `-r` | 100 | Requests per second limit |
| `--count` | `-n` | 1000 | Total number of requests |
| `--timeout` | - | 10s | Request timeout |
| `--method` | `-m` | GET | HTTP method (GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS) |
| `--json` | `-j` | false | Output in JSON format |
| `--stats` | `-s` | - | Save stats to file |

#### Examples

**Basic GET request:**
```bash
portman request -u https://example.com
```

**High RPS load test:**
```bash
portman request -u https://example.com -w 50 -r 1000 -n 50000
```

**HEAD method (fastest):**
```bash
portman request -u https://example.com -m HEAD -w 100 -r 5000 -n 100000
```

---

### Report (`report`)

Generate a comprehensive JSON report.

```bash
portman report
```

#### Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | - | Output file path |
| `--type` | `-t` | health | Report type (health, load-test, combined) |
| `--checks` | `-c` | - | JSON file with check results |
| `--stats` | `-s` | - | JSON file with request statistics |

---

## Configuration

### Config File

Create a `config.yaml` in your home directory, project root, or `$HOME/.config/portman/`:

```yaml
url: ""
workers: 10
rps: 100
timeout: 10
log_level: info
log_file: ""
output_format: "text"
method: "GET"
max_retries: 3
retry_wait_ms: 500
tls_insecure: false

# Custom headers (optional)
headers:
  Authorization: "Bearer token"
  X-Custom-Header: "value"
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `PORTMAN_URL` | Default URL |
| `PORTMAN_WORKERS` | Default workers |
| `PORTMAN_RPS` | Default RPS |
| `PORTMAN_TIMEOUT` | Default timeout |
| `PORTMAN_LOG_LEVEL` | Log level (debug, info, warn, error) |
| `PORTMAN_LOG_FILE` | Log file path |
| `PORTMAN_OUTPUT_FORMAT` | Output format (text, json) |
| `PORTMAN_METHOD` | Default HTTP method |
| `PORTMAN_TLS_INSECURE` | Skip TLS verification |
| `PORTMAN_DB_PATH` | History database path |
| `PORTMAN_WEBHOOK_URL` | Webhook URL for alerts |
| `PORTMAN_MAX_LATENCY` | Max latency threshold |

### .env File

Copy `.env.example` to `.env` and customize:

```bash
cp .env.example .env
```

---

## Webhook Notifications

Portman supports sending alerts to various platforms when endpoints go down or recover.

### Slack

```bash
portman check -u https://example.com --watch --webhook https://hooks.slack.com/services/xxx
```

### Discord

```bash
portman check -u https://example.com --watch --webhook https://discord.com/api/webhooks/xxx
```

### Telegram

```bash
portman check -u https://example.com --watch --webhook https://api.telegram.org/botxxx/sendMessage?chat_id=xxx
```

---

## Safe Testing Thresholds

| VPS RAM | Max Workers (No Cache) | Max Workers (With Cache) |
|---------|----------------------|------------------------|
| 1 GB    | 10–15                | 50–100                 |
| 2 GB    | 20–30                | 100–200                |
| 4 GB    | 40–60                | 300–500                |
| 8 GB    | 80–120               | 600–1,000              |

---

## Use Cases

### Basic Monitoring

```bash
# Single URL check
portman check -u https://example.com

# Multiple URLs
portman check -u https://example.com -u https://api.example.com
```

### Continuous Monitoring with Alerts

```bash
# Watch mode with Slack webhook
portman check -u https://example.com --watch --webhook https://hooks.slack.com/xxx

# With latency threshold
portman check -u https://example.com --watch --max-latency 2s --webhook https://hooks.slack.com/xxx
```

### Scheduled Monitoring

```bash
# Every 5 minutes with alerts
portman schedule --cron "*/5 * * * *" -u https://example.com --webhook https://hooks.slack.com/xxx

# Hourly checks
portman schedule --cron "0 * * * *" -u https://example.com https://api.example.com
```

### SSL Certificate Monitoring

```bash
# Check SSL for multiple URLs
portman ssl -u https://example.com -u https://google.com
```

### Real-time Dashboard

```bash
# Interactive dashboard
portman dashboard -u https://example.com -u https://google.com --interval 5s
```

### Load Testing

```bash
# Basic load test
portman request -u https://example.com -w 50 -r 500 -n 10000

# Stress test with HEAD
portman request -u https://example.com -m HEAD -w 100 -r 5000 -n 50000
```

### History and Analytics

```bash
# Enable history saving
portman check -u https://example.com --save

# View history
portman history

# Get statistics
portman history --stats https://example.com

# Export history
portman history --export -o backup.json
```

---

## Troubleshooting

### Connection Errors

- Check if the URL is correct
- Increase timeout with `--timeout`
- Check firewall/network settings
- Try setting `tls_insecure: true` in config for self-signed certs

### High Failure Rate

- Reduce workers/RPS
- Check target server logs
- Ensure rate limiting on target server
- Check if target server is overwhelmed

### Performance Issues

- Use HEAD method for faster tests
- Increase workers for more concurrency
- Check your network upload speed

---

## License

MIT
