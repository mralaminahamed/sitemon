# History Command

View health check history from SQLite database.

## Usage

```bash
portman history
```

## Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--db` | - | ~/.portman/history.db | Path to history database |
| `--limit` | `-n` | 50 | Number of records to show |
| `--from` | - | - | Start date (YYYY-MM-DD) |
| `--to` | - | - | End date (YYYY-MM-DD) |
| `--stats` | - | - | Show statistics for URL |
| `--export` | - | false | Export all history as JSON |
| `--output` | `-o` | - | Output file for export |

## Examples

### View recent history

```bash
portman history
```

### Show last 100 records

```bash
portman history --limit 100
```

### Show statistics for URL

```bash
portman history --stats https://example.com
```

### Export to JSON

```bash
portman history --export -o history.json
```

### Filter by date

```bash
portman history --from 2026-01-01 --to 2026-02-27
```

## Output

```
=== Health Check History ===

[2026-02-27 13:25] ✓ https://example.com - 200 in 45ms
[2026-02-27 13:20] ✓ https://example.com - 200 in 42ms
[2026-02-27 13:15] ✗ https://example.com - 0 in 0ms
[2026-02-27 13:10] ✓ https://example.com - 200 in 44ms
```

### Statistics Output

```
=== Statistics for https://example.com ===
Total Checks:    100
Successful:      98
Failed:          2
Uptime:          98.00%
Avg Response:    45.23ms
```
