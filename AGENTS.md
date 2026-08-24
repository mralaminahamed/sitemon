# Sitemon — AGENTS.md

## Overview

Full-stack, event-driven site health monitoring platform. Go microservices +
React/TypeScript dashboard, coordinated over NATS, with MongoDB, Redis, and
Claude-powered incident analysis.

- **Language:** Go 1.27 (backend), TypeScript/React 19 (web)
- **Module:** `github.com/mralaminahamed/sitemon`

## Layout

```
apps/
  gateway/    Echo REST + WebSocket, /metrics; routes to services
  checker/    health/ssl/load engine; RPC + check.job consumer
  scheduler/  cron → check.job
  notifier/   check.result → webhooks (up/down transitions)
  ai/         anomaly analysis + Claude summaries (ai.analyze RPC)
  ai/mcp/     stdio MCP server (check_url, ssl_check, analyze)
  cli/        Cobra CLI (check, ssl, request, dashboard, schedule)
  web/        Vite + React 19 + TS dashboard
packages/shared/  http ssl monitor notify scheduler loadtest store cache bus
                  models health logger config validation tui metrics
infra/      docker-compose(.prod).yml, caddy, mongo, prometheus, grafana, k8s
```

Single Go module; each `apps/<svc>/cmd` builds its own binary.

## Conventions

- **Bus:** NATS JetStream for events (`check.job`, `check.result`,
  `alert.raised`); core request-reply for RPC (`checker.run`,
  `checker.loadtest`, `ai.analyze`). All wire formats live in `packages/shared/bus`.
- **Graceful degradation:** gateway runs in-process when no bus; ai/notifier
  no-op cleanly without a key/webhook.
- **Claude:** Anthropic Go SDK, default model `claude-opus-5` (`ANTHROPIC_MODEL`).
- **Comments:** minimal — required, short, no over-explaining.
- **Git:** small scoped Conventional Commits; PR per change; merge (not squash)
  to `trunk`.

## Commands

```bash
make up            # local stack
make build         # all service binaries
make test-race     # go test -race ./...
make lint          # go vet + gofmt check
make obs-up        # prometheus + grafana
go run ./apps/cli check -u https://example.com
```

## Verify before done

`gofmt -l`, `go vet ./...`, `go build ./...`, `CGO_ENABLED=1 go test -race ./...`.
