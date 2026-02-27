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

Output:
```
Portman version dev
  commit: none
  date: unknown
  built by: unknown
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

**With custom timeout:**
```bash
portman check --url https://example.com --timeout 5s
```

**Watch mode (continuous monitoring):**
```bash
portman check --url https://example.com --watch --interval 30s
portman check -u https://example.com -w -i 1m
```

**JSON output:**
```bash
portman check -u https://example.com --json
```

**Save results to file:**
```bash
portman check -u https://example.com -o results.json
```

#### Sample Output

```
[12:59:26] ✓ https://example.com - 200 (UP) in 42.9225ms
[12:59:28] ✓ https://google.com - 200 (UP) in 1.581973375s

--- Summary ---
Total Checks: 2
Successful: 2
Failed: 0
Uptime: 100.00%
Avg Response Time: 811ms
Min Response Time: 42.9225ms
Max Response Time: 1.581973375s
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

**Fastest method (HEAD - no body download):**
```bash
portman request -u https://example.com -m HEAD -w 100 -r 5000 -n 100000
```

**POST request:**
```bash
portman request -u https://example.com/api/submit -m POST -w 20 -r 200 -n 5000
```

**PUT request:**
```bash
portman request -u https://example.com/api/update -m PUT -w 10 -r 100 -n 1000
```

**PATCH request:**
```bash
portman request -u https://example.com/api/patch -m PATCH -w 10 -r 100 -n 1000
```

**DELETE request:**
```bash
portman request -u https://example.com/api/resource/123 -m DELETE -w 10 -r 100 -n 1000
```

**OPTIONS request:**
```bash
portman request -u https://example.com -m OPTIONS -n 100
```

**JSON output:**
```bash
portman request -u https://example.com --json
```

**Save stats to file:**
```bash
portman request -u https://example.com -s stats.json
```

#### Sample Output

```
======================================
Target URL:     https://example.com
Method:        GET
Workers:       10
RPS Limit:     100
--------------------------------------
Total Requests: 1000
Successful:     998 (99.80%)
Failed:         2
Duration:       10.234s
Requests/sec:   97.72
--------------------------------------
Latency Stats (ms):
  Min:    11.41ms
  Avg:    18ms
  Max:    53.74ms
  P50:    14.94ms
  P90:    40.95ms
  P95:    42.70ms
  P99:    53.74ms
======================================
```

---

### Report (`report`)

Generate a comprehensive JSON report for health checks or load tests.

```bash
portman report
```

#### Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | - | Output file path (default is stdout) |
| `--type` | `-t` | health | Report type (health, load-test, combined) |
| `--checks` | `-c` | - | JSON file with check results |
| `--stats` | `-s` | - | JSON file with request statistics |

#### Examples

**Output to file:**
```bash
portman report -o report.json
```

**Output to stdout:**
```bash
portman report
```

**Generate from check results:**
```bash
portman check -u https://example.com -o checks.json
portman report --checks checks.json -o report.json
```

**Generate from request stats:**
```bash
portman request -u https://example.com -s stats.json
portman report --stats stats.json -o report.json
```

**Combined report:**
```bash
portman report --checks checks.json --stats stats.json -o combined.json
```

#### Sample Output

```json
{
  "generated_at": "2026-02-27T12:00:00Z",
  "report_type": "health",
  "summary": {
    "total_checks": 100,
    "successful": 98,
    "failed": 2,
    "uptime_percentage": 98
  },
  "checks": [
    {
      "url": "https://example.com",
      "status": "UP",
      "status_code": 200,
      "response_time_ms": 45000000,
      "timestamp": "2026-02-27T12:00:01Z"
    }
  ]
}
```

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

### Log to File

```bash
portman check --url https://example.com --log-file /tmp/portman.log
```

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

# Multiple URLs
portman check -u https://yourportfolio.com -u https://blog.yourportfolio.com

# Continuous monitoring
portman check --url https://yourportfolio.com --watch --interval 5m

# Save results for later analysis
portman check -u https://yourportfolio.com -o health_results.json
```

### Load Testing

```bash
# Light load test
portman request -u https://yourportfolio.com -w 20 -r 100 -n 1000

# Medium load test
portman request -u https://yourportfolio.com -w 50 -r 500 -n 10000

# Heavy load test
portman request -u https://yourportfolio.com -w 100 -r 1000 -n 50000

# Extreme stress test (use with caution)
portman request -u https://yourportfolio.com -m HEAD -w 200 -r 5000 -n 100000

# Save detailed stats
portman request -u https://yourportfolio.com -s load_test_stats.json
```

### API Testing

```bash
# Test API health endpoint
portman request -u https://api.example.com/health -m GET

# Test POST endpoint
portman request -u https://api.example.com/users -m POST -w 10 -r 50

# Test PUT endpoint
portman request -u https://api.example.com/users/123 -m PUT -w 10 -r 50

# Test PATCH endpoint
portman request -u https://api.example.com/users/123 -m PATCH -w 10 -r 50

# Test DELETE endpoint
portman request -u https://api.example.com/users/123 -m DELETE -w 5 -r 20

# Check allowed methods
portman request -u https://api.example.com/endpoint -m OPTIONS
```

### Generate Reports

```bash
# Run health checks and save
portman check -u https://example.com -u https://google.com -o checks.json

# Run load test and save stats
portman request -u https://example.com -s stats.json

# Generate health report
portman report --checks checks.json -o health_report.json

# Generate load test report
portman report --stats stats.json -o load_report.json
```

---

## Troubleshooting

### Connection Errors

If you see connection errors:
- Check if the URL is correct
- Increase timeout with `--timeout`
- Check firewall/network settings
- Try `--tls-insecure` if using self-signed cert

### High Failure Rate

If seeing many failures:
- Reduce workers/RPS
- Check target server logs
- Ensure rate limiting on target server
- Check if target server is overwhelmed

### Performance Issues

If tool is slow:
- Use HEAD method for faster tests
- Increase workers for more concurrency
- Check your network upload speed

### TLS Certificate Errors

If you get certificate errors:
```bash
# In config.yaml
tls_insecure: true

# Or use environment variable
export PORTMAN_TLS_INSECURE=true
```

---

## License

MIT
