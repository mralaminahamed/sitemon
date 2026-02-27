# Portman Documentation

## Overview

Portman is a CLI tool for HTTP request management and portfolio site health monitoring. It provides health checks, continuous monitoring, high-volume request testing, and JSON report generation.

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
Status: UP
Status Code: 200
Response Time: 43.232042ms
Checked at: 2026-02-27 12:32:59.694309 +0600 +06 m=+0.044201460
```

---

## Commands

### Health Check

Check the health status of a URL:

```bash
portman check --url https://example.com
```

#### Examples

**Basic check:**
```bash
portman check -u https://example.com
```

**With custom timeout:**
```bash
portman check --url https://example.com --timeout 5s
```

**Watch mode (continuous monitoring):**
```bash
portman check --url https://example.com --watch --interval 30s
```

**Watch with custom interval:**
```bash
portman check -u https://example.com -w -i 1m
```

Output:
```
Watching https://example.com every 30s (Ctrl+C to stop)
[12:32:59] https://example.com - 200 (UP) in 43.23ms
[12:33:29] https://example.com - 200 (UP) in 41.12ms
[12:33:59] https://example.com - 200 (UP) in 44.56ms
```

---

### Request (Load Testing)

Send high-volume HTTP requests for load testing:

```bash
portman request --url https://example.com --workers 100 --rps 500 --count 10000
```

#### Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--url` | `-u` | - | Target URL (required) |
| `--workers` | `-w` | 10 | Number of concurrent workers |
| `--rps` | `-r` | 100 | Requests per second limit |
| `--count` | `-n` | 1000 | Total number of requests |
| `--timeout` | - | 10s | Request timeout |
| `--method` | `-m` | GET | HTTP method (GET, HEAD, POST, DELETE) |

#### Examples

**Basic GET request:**
```bash
portman request -u https://example.com
```

**High RPS load test:**
```bash
portman request -u https://example.com -w 50 -r 1000 -n 50000
```

**Fastest method (HEAD - no body download):**
```bash
portman request -u https://example.com -m HEAD -w 100 -r 5000 -n 100000
```

**POST request:**
```bash
portman request -u https://example.com/api/submit -m POST -w 20 -r 200 -n 5000
```

**DELETE request:**
```bash
portman request -u https://example.com/api/resource/123 -m DELETE -w 10 -r 100 -n 1000
```

**Custom timeout:**
```bash
portman request -u https://example.com --timeout 5s -n 5000
```

**Sample Output:**
```
[====================================================================] 1000/1000 (100%)
--- Results ---
Total Requests: 1000
Successful: 998
Failed: 2
Duration: 10.234s
Requests/sec: 97.72
Avg Latency: 45.23ms
```

---

### Report

Generate a JSON health report:

```bash
portman report --output report.json
```

#### Examples

**Output to file:**
```bash
portman report -o report.json
```

**Output to stdout:**
```bash
portman report
```

**Sample Output:**
```json
{
  "generated_at": "2026-02-27",
  "total_checks": 0,
  "successful": 0,
  "failed": 0,
  "average_latency": "0ms",
  "uptime_percentage": "0%"
}
```

---

## Configuration

### Config File

Create a `config.yaml` in your home directory or project root:

```yaml
url: ""
workers: 10
rps: 100
timeout: 10
log_level: info
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `PORTMAN_URL` | Default URL |
| `PORTMAN_WORKERS` | Default workers |
| `PORTMAN_RPS` | Default RPS |
| `PORTMAN_TIMEOUT` | Default timeout |
| `PORTMAN_LOG_LEVEL` | Log level (debug, info, warn, error) |

### CLI Flags

Global flags available for all commands:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | config.yaml | Config file path |
| `--verbose` | `-v` | false | Enable verbose output |
| `--log-level` | - | info | Log level |

---

## Safe Testing Thresholds

### VPS Memory Limits

| VPS RAM | Max Workers (No Cache) | Max Workers (With Cache) |
|---------|----------------------|------------------------|
| 1 GB    | 10–15                | 50–100                 |
| 2 GB    | 20–30                | 100–200                |
| 4 GB    | 40–60                | 300–500                |
| 8 GB    | 80–120               | 600–1,000              |

### Performance Tips

1. **Use HEAD method** - Fastest, doesn't download response body
2. **Increase workers** - More concurrent connections
3. **Match RPS to network** - Don't exceed your upload bandwidth
4. **Monitor target server** - Watch for CPU/memory saturation

---

## Logging

### Log Levels

- `debug` - Detailed debug information
- `info` - Normal operation (default)
- `warn` - Warning messages
- `error` - Error messages only

### Example

```bash
portman check --url https://example.com --log-level debug
```

---

## Use Cases

### Portfolio Site Monitoring

```bash
# Daily health check
portman check --url https://yourportfolio.com

# Continuous monitoring
portman check --url https://yourportfolio.com --watch --interval 5m
```

### Load Testing

```bash
# Light load test
portman request -u https://yourportfolio.com -w 20 -r 100 -n 1000

# Heavy load test
portman request -u https://yourportfolio.com -w 100 -r 1000 -n 50000

# Extreme stress test (use with caution)
portman request -u https://yourportfolio.com -m HEAD -w 200 -r 5000 -n 100000
```

### API Testing

```bash
# Test API endpoint
portman request -u https://api.example.com/health -m GET

# Test POST endpoint
portman request -u https://api.example.com/users -m POST -w 10 -r 50

# Test DELETE endpoint
portman request -u https://api.example.com/users/123 -m DELETE -w 5 -r 20
```

---

## Troubleshooting

### Connection Errors

If you see connection errors:
- Check if the URL is correct
- Increase timeout with `--timeout`
- Check firewall/network settings

### High Failure Rate

If seeing many failures:
- Reduce workers/RPS
- Check target server logs
- Ensure rate limiting on target server

### Performance Issues

If tool is slow:
- Use HEAD method for faster tests
- Increase workers for more concurrency
- Check your network upload speed
