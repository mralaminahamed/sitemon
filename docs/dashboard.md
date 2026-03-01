# Dashboard Command

Run interactive TUI dashboard for real-time monitoring.

## Usage

```bash
sitemon dashboard --url https://example.com
```

## Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--url` | `-u` | [] | URL(s) to monitor |
| `--interval` | - | 5s | Refresh interval |
| `--timeout` | - | 10s | Request timeout |

## Examples

### Single URL

```bash
sitemon dashboard -u https://example.com
```

### Multiple URLs

```bash
sitemon dashboard -u https://example.com -u https://google.com
```

### Custom interval

```bash
sitemon dashboard -u https://example.com --interval 10s
```

### With Cloudflare bypass

```bash
sitemon dashboard -u https://codecept.io --bypass-cloudflare
```

## Features

- **Real-time monitoring** - Continuously checks all URLs at specified interval
- **Full terminal width** - Dynamically adjusts to terminal size
- **Recent activity** - Shows last 10 check results
- **Statistics** - Total, UP, DOWN counts with uptime percentage
- **Latency tracking** - Average, min, max response times
- **RPS tracking** - Requests per second for each URL

## Controls

- **Ctrl+C** - Exit the dashboard (shows goodbye message)

## Output

```
 ◈ Sitemon Dashboard 
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 URLs: 2 | Interval: 5s | Running: 2m30s | Updated: 13:25
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ┌────────┐ ┌────────┐ ┌────────┐ ┌──────────────┐
  │ TOTAL 2│ │ UP 2   │ │ DOWN 0 │ │ UPTIME 100% │
  └────────┘ └────────┘ └────────┘ └──────────────┘
  ┌────────┐ ┌────────┐ ┌────────┐ ┌──────────────┐
  │ AVG 45ms│ │ MIN 32ms│ │ MAX120ms│ │ RPS 0.2    │
  └────────┘ └────────┘ └────────┘ └──────────────┘

 #  URL                                        STATUS         CODE        LATENCY
 ──────────────────────────────────────────────────────────────────────────────────
 1  https://example.com                       ● UP           200         45ms
 2  https://google.com                        ● UP           200         120ms

 Recent Activity 
────────────────────────────────────────────────────────────────────────────
 [13:25] ✓ https://example.com - 200 (UP) in 45ms
 [13:25] ✓ https://google.com - 200 (UP) in 120ms
 [13:20] ✓ https://example.com - 200 (UP) in 42ms
 [13:20] ✓ https://google.com - 200 (UP) in 118ms
 [13:15] ✓ https://example.com - 200 (UP) in 44ms
 [13:15] ✓ https://google.com - 200 (UP) in 115ms
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 Ctrl+C to exit | Interval: 5s | Uptime: 2m30s
```

## Status Indicators

- **● UP** (green) - Server is responding normally
- **✗ DOWN** (red) - Server is down or unreachable
- **⚠ WARN** (yellow) - Server returned 4xx status code
- **↪ REDIRECT** (cyan) - Server returned 3xx redirect

## Latency Colors

- **Green** - Response time < 200ms
- **Cyan** - Response time < 500ms
- **Yellow** - Response time < 1000ms
- **Red** - Response time >= 1000ms
