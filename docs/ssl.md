# SSL Certificate Command

Check SSL certificate details for URLs.

## Usage

```bash
sitemon ssl [url...]
```

## Options

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--url` | `-u` | [] | URL(s) to check |
| `--timeout` | - | 10s | Request timeout |

## Examples

### Check single URL

```bash
sitemon ssl -u https://example.com
```

### Check multiple URLs

```bash
sitemon ssl -u https://example.com -u https://google.com
```

## Output

```
=== SSL Certificate: https://example.com ===
Status:     ✓ Valid
Issuer:     SSL Corporation
Subject:    example.com
Valid From: 2026-02-13 18:53:48
Valid Until: 2026-05-14 18:57:50
Days Left:  76 days
Protocol:   TLS 772
Cipher:     0x1301
```

## Status Indicators

- **✓ Valid** - Certificate is valid (more than 30 days remaining)
- **⚠ Warning** - Certificate expires soon (7-30 days remaining)
- **✗ Critical** - Certificate expires very soon (less than 7 days)
