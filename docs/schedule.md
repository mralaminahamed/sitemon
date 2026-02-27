# Schedule Command

Run health checks on a cron schedule.

## Usage

```bash
portman schedule --cron "*/5 * * * *" --url https://example.com
```

## Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--cron` | - | - | Cron expression (required) |
| `--url` | `-u` | [] | URL(s) to check |
| `--timeout` | - | 10s | Request timeout |
| `--webhook` | - | - | Webhook URL for alerts |

## Cron Expression Examples

| Expression | Description |
|------------|-------------|
| `*/5 * * * *` | Every 5 minutes |
| `*/15 * * * *` | Every 15 minutes |
| `*/30 * * * *` | Every 30 minutes |
| `0 * * * *` | Every hour |
| `0 */2 * * *` | Every 2 hours |
| `0 0 * * *` | Daily at midnight |
| `0 9 * * *` | Daily at 9 AM |
| `0 9 * * 1-5` | Weekdays at 9 AM |

## Examples

### Every 5 minutes

```bash
portman schedule --cron "*/5 * * * *" -u https://example.com
```

### Every hour with webhook

```bash
portman schedule --cron "0 * * * *" -u https://example.com --webhook https://hooks.slack.com/xxx
```

### Multiple URLs

```bash
portman schedule --cron "*/15 * * * *" -u https://example.com -u https://api.example.com
```

### Daily at 9 AM

```bash
portman schedule --cron "0 9 * * *" -u https://example.com
```

## Output

```
Portman Scheduler
=================
Schedule: */5 * * * * (every 5 minutes)
URLs: [https://example.com]
Timeout: 10s
Webhook: https://hooks.slack.com/xxx

Press Ctrl+C to stop

Next run at: Fri, 27 Feb 2026 12:05:00 +06

[12:00:00] Running health checks...
  ✓ https://example.com - 200 (UP) in 45ms

Next run at: Fri, 27 Feb 2026 12:05:00 +06
```
