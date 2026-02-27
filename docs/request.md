# Request Command

Send high-volume HTTP requests for load testing.

## Usage

```bash
portman request [url]
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
portman request -u https://example.com
```

### High RPS load test

```bash
portman request -u https://example.com -w 50 -r 1000 -n 50000
```

### HEAD method (fastest)

```bash
portman request -u https://example.com -m HEAD -w 100 -r 5000 -n 100000
```

### POST request

```bash
portman request -u https://api.example.com/submit -m POST -w 20 -r 200 -n 5000
```

### DELETE request

```bash
portman request -u https://api.example.com/resource/123 -m DELETE -w 10 -r 100 -n 1000
```

### Save stats to file

```bash
portman request -u https://example.com -s stats.json
```

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
