# Health Check Command

Check the health status of one or more URLs.

## Usage

```bash
portman check [url...]
```

## Options

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

## Examples

### Basic check

```bash
portman check -u https://example.com
```

### Check multiple URLs

```bash
portman check -u https://example.com -u https://google.com
portman check https://example.com https://google.com
```

### Watch mode

```bash
portman check -u https://example.com --watch --interval 30s
```

### With webhook alerts

```bash
portman check -u https://example.com --watch --webhook https://hooks.slack.com/services/xxx
```

### Content validation

```bash
portman check -u https://example.com --contains "Welcome" --not-contains "Error"
```

### Check SSL certificate

```bash
portman check -u https://example.com --check-ssl
```

### Save to history

```bash
portman check -u https://example.com --save
```

## Output

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
