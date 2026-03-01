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
| `--timeout` | - | 10s | Request timeout |
| `--method` | `-m` | GET | HTTP method (GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS) |
| `--json` | `-j` | false | Output in JSON format |
| `--stats` | `-s` | - | Save stats to file |
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

### Extreme load test (1M with Cloudflare bypass)

```bash
sitemon request -u https://codecept.io -w 100 -r 1000 -n 1000000 --bypass-cloudflare
```

### HEAD method (fastest, ideal for millions)

```bash
sitemon request -u https://example.com -m HEAD -w 200 -r 5000 -n 1000000
```

### POST request with data

```bash
sitemon request -u https://api.example.com/submit -m POST -w 20 -r 200 -n 5000
```

### DELETE request

```bash
sitemon request -u https://api.example.com/resource/123 -m DELETE -w 10 -r 100 -n 1000
```

### Save stats to file

```bash
sitemon request -u https://example.com -s stats.json
```

### Bypass Cloudflare

```bash
sitemon request -u https://codecept.io -w 10 -r 100 --bypass-cloudflare
```

### Save million request results

```bash
sitemon request -u https://example.com -w 200 -r 2000 -n 1000000 -s million_stats.json
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

## Performance Tips

- Use **HEAD** method for fastest results (no body download)
- Increase **workers** for more concurrent connections
- Match **rps** to your network bandwidth
- Monitor target server resources
- For millions of requests, use `--stats` to save results
- Use `--bypass-cloudflare` for Cloudflare-protected sites
