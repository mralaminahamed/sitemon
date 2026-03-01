# Request Command

Send high-volume HTTP requests for load testing.

## Usage

```bash
sitemon request [url]
```

## Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--url` | `-u` | - | Target URL (required) |
| `--workers` | `-w` | 10 | Number of concurrent workers |
| `--rps` | `-r` | 100 | Requests per second limit |
| `--count` | `-n` | 1000 | Total number of requests |
| `--duration` | - | - | Run for duration (e.g., 30s, 5m) |
| `--ramp-up` | - | - | Ramp-up period (e.g., 10s) |
| `--timeout` | - | 10s | Request timeout |
| `--method` | `-m` | GET | HTTP method (GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS) |
| `--body` | - | - | Request body for POST/PUT/PATCH (JSON string) |
| `--header` | `-H` | - | Custom header (key:value) - can be used multiple times |
| `--cookie` | - | - | Cookie string (name=value) |
| `--proxy` | - | - | HTTP proxy URL |
| `--follow-redirects` | - | false | Follow HTTP redirects |
| `--json` | `-j` | false | Output in JSON format |
| `--csv` | - | false | Output in CSV format |
| `--prometheus` | - | false | Output in Prometheus format |
| `--junit` | - | false | Output in JUnit XML format |
| `--stats` | `-s` | - | Save stats to file |
| `--quiet` | - | false | Suppress progress bar |
| `--bypass-cloudflare` | - | false | Use browser headers to bypass Cloudflare bot detection |

## HTTP Methods

| Method | Description |
|--------|-------------|
| GET | Default, retrieves content |
| HEAD | Fastest, no response body |
| POST | Send data to endpoint |
| PUT | Update resource |
| PATCH | Partial update |
| DELETE | Remove resource |
| OPTIONS | Check allowed methods |

## Examples

### Basic GET request

```bash
sitemon request -u https://example.com
```

### Quick load test (10K requests)

```bash
sitemon request -u https://example.com -w 50 -r 500 -n 10000
```

### Moderate load test (100K requests)

```bash
sitemon request -u https://example.com -w 100 -r 1000 -n 100000
```

### Heavy load test (1M requests)

```bash
sitemon request -u https://example.com -w 200 -r 2000 -n 1000000
```

### Duration-based testing

```bash
# Run for 30 seconds
sitemon request -u https://example.com -w 50 -r 500 --duration 30s

# Run for 5 minutes
sitemon request -u https://example.com -w 100 -r 1000 --duration 5m

# Run for 1 hour
sitemon request -u https://example.com -w 200 -r 2000 --duration 1h
```

### Ramp-up period

```bash
# Gradually increase RPS from 0 to 1000 over 30 seconds
sitemon request -u https://example.com -w 50 -r 1000 -n 50000 --ramp-up 30s
```

### With custom headers

```bash
# Single header
sitemon request -u https://api.example.com -H "Authorization: Bearer token"

# Multiple headers
sitemon request -u https://api.example.com -H "Authorization: Bearer token" -H "X-Custom: value"
```

### With cookies

```bash
sitemon request -u https://example.com --cookie "session=abc123; user=john"
```

### With request body (POST)

```bash
sitemon request -u https://api.example.com/login -m POST --body '{"username":"admin","password":"secret"}'
```

### Through proxy

```bash
sitemon request -u https://example.com --proxy http://proxy:8080
```

### Follow redirects

```bash
sitemon request -u http://example.com --follow-redirects
```

### CSV output

```bash
sitemon request -u https://example.com -n 10000 --csv
```

### Prometheus metrics output

```bash
sitemon request -u https://example.com -n 10000 --prometheus
```

### JUnit XML output (for CI/CD)

```bash
sitemon request -u https://example.com -n 1000 --junit
```

### Quiet mode (no progress bar)

```bash
sitemon request -u https://example.com -n 10000 --quiet
```

### Cloudflare bypass

```bash
sitemon request -u https://codecept.io -w 10 -r 100 --bypass-cloudflare
```

### Combined options

```bash
sitemon request -u https://api.example.com/secure \
  -m POST \
  --body '{"key":"value"}' \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -w 50 -r 500 -n 10000 \
  --duration 1m \
  --stats results.json
```

## Flood Testing (High RPS)

Test your server's ability to handle high request volumes.

### Aggressive Load Tests

| RPS | Workers | Requests | Use Case |
|-----|---------|----------|----------|
| 1,000 | 50 | 10,000 | Moderate traffic spike |
| 2,000 | 100 | 50,000 | Heavy load test |
| 5,000 | 200 | 100,000 | Stress testing |
| 10,000 | 500 | 500,000 | DDoS simulation |
| 20,000 | 1000 | 1,000,000 | Extreme stress test |

### Flood Testing Examples

```bash
# 1K RPS flood test
sitemon request -u https://example.com -w 50 -r 1000 -n 10000

# 5K RPS flood test
sitemon request -u https://example.com -w 200 -r 5000 -n 50000

# 10K RPS stress test
sitemon request -u https://example.com -w 500 -r 10000 -n 100000

# 20K RPS extreme test (1M requests)
sitemon request -u https://example.com -w 1000 -r 20000 -n 1000000

# 10K RPS with HEAD (fastest possible)
sitemon request -u https://example.com -m HEAD -w 500 -r 10000 -n 1000000

# 5K RPS with Cloudflare bypass
sitemon request -u https://codecept.io -w 200 -r 5000 -n 50000 --bypass-cloudflare
```

### API Endpoint Flooding

```bash
# POST flood to login endpoint
sitemon request -u https://api.example.com/login -m POST -w 100 -r 1000 -n 10000

# GET flood to specific endpoint
sitemon request -u https://api.example.com/users -m GET -w 200 -r 5000 -n 50000

# JSON payload POST flood
sitemon request -u https://api.example.com/submit -m POST -w 100 -r 2000 -n 20000
```

### Continuous Flood Mode

```bash
# Run for 1 hour at 5K RPS (18M requests)
sitemon request -u https://example.com -w 500 -r 5000 --duration 1h

# Sustained 10K RPS attack simulation
sitemon request -u https://example.com -w 1000 -r 10000 --duration 10m
```

## Load Testing Guide

### Safe Thresholds by VPS RAM

| VPS RAM | Max Safe Workers (No Cache) | Max Safe Workers (With Cache) |
|---------|------------------------------|-------------------------------|
| 1 GB | 10-15 | 50-100 |
| 2 GB | 20-30 | 100-200 |
| 4 GB | 40-60 | 300-500 |
| 8 GB | 80-120 | 600-1,000 |

### Recommended Configurations

| Request Count | Workers | RPS | Estimated Time |
|---------------|---------|-----|----------------|
| 1,000 | 10 | 100 | ~10s |
| 10,000 | 50 | 500 | ~20s |
| 100,000 | 100 | 1,000 | ~1.5min |
| 1,000,000 | 200 | 2,000 | ~8min |
| 1,000,000 (HEAD) | 200 | 5,000 | ~3min |

## Output

### Standard Output

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
Latency Stats:
  Min:    11.41ms
  Avg:    18ms
  Max:    53.74ms
  P50:    14.94ms
  P90:    40.95ms
  P95:    42.70ms
  P99:    53.74ms
--------------------------------------
Response Size:
  Avg:    1250 bytes
  Min:    800 bytes
  Max:    5000 bytes
--------------------------------------
Status Codes:
  200: 995
  301: 3
  404: 2
======================================
```

### JSON Output

```json
{
  "total_requests": 1000,
  "successful": 998,
  "failed": 2,
  "duration": "10.234s",
  "requests_per_sec": 97.72,
  "success_rate": 99.8,
  "avg_latency_ms": "18ms",
  "min_latency_ms": "11.41ms",
  "max_latency_ms": "53.74ms",
  "p50_latency_ms": "14.94ms",
  "p90_latency_ms": "40.95ms",
  "p95_latency_ms": "42.70ms",
  "p99_latency_ms": "53.74ms",
  "target_url": "https://example.com",
  "method": "GET",
  "workers": 10,
  "rps_limit": 100,
  "response_size_avg_bytes": 1250,
  "status_codes": {
    "200": 995,
    "301": 3,
    "404": 2
  }
}
```

### Prometheus Output

```
# HELP sitemon_total_requests Total number of requests
# TYPE sitemon_total_requests counter
sitemon_total_requests 1000
# HELP sitemon_successful Successful requests
# TYPE sitemon_successful counter
sitemon_successful 998
# HELP sitemon_failed Failed requests
# TYPE sitemon_failed counter
sitemon_failed 2
# HELP sitemon_requests_per_second Requests per second
# TYPE sitemon_requests_per_second gauge
sitemon_requests_per_second 97.72
# HELP sitemon_avg_latency_ms Average latency in milliseconds
# TYPE sitemon_avg_latency_ms gauge
sitemon_avg_latency_ms 18.00
# HELP sitemon_p99_latency_ms P99 latency in milliseconds
# TYPE sitemon_p99_latency_ms gauge
sitemon_p99_latency_ms 53.74
```

## Performance Tips

- Use **HEAD** method for fastest results (no body download)
- Increase **workers** for more concurrent connections
- Match **rps** to your network bandwidth
- Monitor target server resources
- For millions of requests, use `--stats` to save results
- Use `--bypass-cloudflare` for Cloudflare-protected sites

## Important Notes

### Responsible Usage

Only test servers you own or have permission to test. Unauthorized testing may be illegal.

### Your Server Protection

When flood testing your own servers:

- Monitor server CPU, memory, and network bandwidth
- Start with lower RPS and gradually increase
- Use `--timeout` to detect stalled connections
- Check results for high failure rates indicating server overload

### Network Limits

| Network Type | Max Safe RPS |
|--------------|--------------|
| Home DSL | 500-1,000 |
| Home Fiber | 2,000-5,000 |
| VPS 1Gbps | 10,000-20,000 |
| Dedicated Server | 50,000+ |
