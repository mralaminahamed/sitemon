# Configuration

Sitemon can be configured via config file, environment variables, or `.env` file.

## Config File

Create `config.yaml` in one of these locations:
- Current directory
- `$HOME/.sitemon/`
- `$HOME/.config/sitemon/`

### Example Config

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

# Custom headers
headers:
  Authorization: "Bearer token"
  X-Custom-Header: "value"
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SITEMON_URL` | - | Default URL |
| `SITEMON_WORKERS` | 10 | Default workers |
| `SITEMON_RPS` | 100 | Default RPS |
| `SITEMON_TIMEOUT` | 10 | Default timeout (seconds) |
| `SITEMON_LOG_LEVEL` | info | Log level (debug, info, warn, error) |
| `SITEMON_LOG_FILE` | - | Log file path |
| `SITEMON_OUTPUT_FORMAT` | text | Output format (text, json) |
| `SITEMON_METHOD` | GET | Default HTTP method |
| `SITEMON_TLS_INSECURE` | false | Skip TLS verification |
| `SITEMON_DB_PATH` | ~/.sitemon/history.db | History database path |
| `SITEMON_WEBHOOK_URL` | - | Webhook URL for alerts |
| `SITEMON_MAX_LATENCY` | - | Max latency threshold |

## .env File

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

Edit the values:

```env
SITEMON_URL=
SITEMON_WORKERS=10
SITEMON_RPS=100
SITEMON_TIMEOUT=10
SITEMON_LOG_LEVEL=info
SITEMON_LOG_FILE=
SITEMON_WEBHOOK_URL=
SITEMON_TLS_INSECURE=false
```

## Priority

Configuration is loaded in this order (later overrides earlier):

1. Default values
2. Config file (`config.yaml`)
3. Environment variables
4. CLI flags

## TLS/SSL Options

### Skip TLS Verification

For servers with self-signed certificates:

```bash
# Via config
tls_insecure: true

# Via environment
export SITEMON_TLS_INSECURE=true

# Via CLI (not available, use config)
```

## Logging

### Log Levels

- **debug** - Detailed debug information
- **info** - Normal operation (default)
- **warn** - Warning messages
- **error** - Error messages only

### Log to File

```bash
sitemon check --url https://example.com --log-file /tmp/sitemon.log
```
