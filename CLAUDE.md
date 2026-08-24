# CLAUDE.md

Guidance for Claude Code (and other agents) working in this repo.

## What this is

Sitemon — a full-stack, event-driven site health monitoring platform. Go
microservices coordinated over NATS, a React/TypeScript dashboard, MongoDB +
Redis, Prometheus, and Claude-powered incident analysis with an MCP server.

- Go 1.27, single module `github.com/mralaminahamed/sitemon`
- Web: React 19 + TypeScript (Vite), in `apps/web`

## Layout

```
apps/{gateway,checker,scheduler,notifier,ai,cli,web}   # ai/mcp = MCP server
packages/shared/   # http ssl monitor notify scheduler loadtest store cache bus
                   # models health logger config validation urlguard auth-side metrics
infra/             # docker-compose(.prod).yml, caddy, mongo, prometheus, grafana, k8s
```

Each `apps/<svc>/cmd` builds its own binary.

## Commands

```bash
make up            # local stack (mongo, redis, nats, all services, web)
make build         # all binaries into ./bin
make test-race     # CGO_ENABLED=1 go test -race ./...
make lint          # go vet + gofmt check
make obs-up        # prometheus + grafana
go run ./apps/cli check -u https://example.com
```

## Verify before claiming done

```bash
gofmt -l apps packages && go vet ./... && go build ./... && CGO_ENABLED=1 go test -race ./...
```

Always verify live where it matters (run the service, curl the endpoint) — this
project has been built with live checks at every step.

## Architecture notes

- **Bus:** NATS JetStream for events (`check.job`, `check.result`,
  `alert.raised`); core request-reply for RPC (`checker.run`,
  `checker.loadtest`, `ai.analyze`). All wire formats live in
  `packages/shared/bus`. RPC replies use an error envelope — errors surface as
  Go errors, not zero values. Consumers set MaxDeliver + drop malformed
  payloads (`bus.ErrDrop`).
- **Graceful degradation:** the gateway runs in-process when no bus is present;
  ai/notifier no-op cleanly without a key/webhook.
- **Health vs readiness:** `/health` is liveness (always ok); `/ready` runs
  dependency pings and returns 503 when a dep is down. k8s uses both.
- **Security:** `/api` takes an optional API key (`GATEWAY_API_KEY`); target
  URLs pass through `packages/shared/urlguard` (blocks SSRF to
  private/loopback/link-local); load tests are capped.
- **Claude:** Anthropic Go SDK, default model `claude-opus-5`
  (`ANTHROPIC_MODEL` overrides). Before touching any Claude/Anthropic code,
  consult the `claude-api` guidance — model ids and SDK shapes drift.

## Conventions

- **Comments:** minimal — only what's required, short, no over-explaining.
- **Git:** small single-scope Conventional Commits; branch from `trunk`; open a
  PR per change; merge with a merge commit (not squash) so scoped commits stay
  in history; delete the branch on merge.
  - Types `feat fix docs refactor perf test build ci chore` ·
    scopes `gateway checker scheduler notifier ai web cli infra`.
- **Never commit** `.env` (gitignored). `PLAN.md` and `ARCHITECTURE.md` are kept
  **untracked** via `.git/info/exclude` — do not add them.
- Match the surrounding code's style; reuse `packages/shared` rather than
  duplicating http/ssl/monitor logic.
