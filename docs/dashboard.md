# Dashboard Command

Run interactive TUI dashboard for real-time monitoring.

## Usage

```bash
portman dashboard --url https://example.com
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
portman dashboard -u https://example.com
```

### Multiple URLs

```bash
portman dashboard -u https://example.com -u https://google.com
```

### Custom interval

```bash
portman dashboard -u https://example.com --interval 10s
```

## Output

```
╔══════════════════════════════════════════════════════════════════╗
║                     Portman Health Monitor                      ║
╠══════════════════════════════════════════════════════════════════╣
║ URLs: 2 | Interval: 5s | Time: 13:25:45 ║
╚══════════════════════════════════════════════════════════════════╝

URL                                       Status          Code    Latency
─────────────────────────────────────────────────────────────────────────────
https://example.com                      ✓ UP            200     45.123ms
https://google.com                       ✓ UP            200     120.456ms

Press Ctrl+C to stop...
```

## Controls

- **Ctrl+C** - Exit the dashboard
