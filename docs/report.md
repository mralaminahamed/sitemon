# Report Command

Generate a comprehensive JSON report.

## Usage

```bash
portman report
```

## Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | - | Output file path |
| `--type` | `-t` | health | Report type (health, load-test, combined) |
| `--checks` | `-c` | - | JSON file with check results |
| `--stats` | `-s` | - | JSON file with request statistics |

## Examples

### Output to file

```bash
portman report -o report.json
```

### Generate from check results

```bash
portman check -u https://example.com -o checks.json
portman report --checks checks.json -o report.json
```

### Generate from request stats

```bash
portman request -u https://example.com -s stats.json
portman report --stats stats.json -o report.json
```

### Combined report

```bash
portman report --checks checks.json --stats stats.json -o combined.json
```

## Output Format

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

## Report Types

- **health** - Health check results
- **load-test** - Load test statistics
- **combined** - Both health and load test data
