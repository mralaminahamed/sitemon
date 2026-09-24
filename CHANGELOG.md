# Changelog

All notable changes to this project are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/). The project has no tagged
releases, so entries are grouped by the date their pull requests merged to `trunk`.

## [Unreleased]

### Changed

- Move local dev ports to the 8100 block (#55)

## 2026-08-31

### Changed

- Bump `actions/setup-go` from 5 to 7 (#40)
- Bump `docker/metadata-action` from 5 to 6 (#41)
- Bump `@tanstack/react-query` from 5.102.3 to 5.102.4 in `apps/web` (#42)

## 2026-08-25: hardening and observability

### Added

- `/metrics` on every service, plus business metrics (#36)
- Grafana dashboard and Prometheus alert rules (#39)
- CI jobs for the web app and govulncheck, Dependabot config, non-root images (#27)

### Fixed

- Check monitors added through the API: scheduler `MONGO_URI` and an immediate first check (#19)
- Self-host web fonts instead of loading them from the Google Fonts CDN (#20)
- Reliable alerting in the notifier: no lost alerts, flap damping, alert history (#21)
- Monitor lifecycle bugs in the gateway: delete, URL normalization, limits (#22)
- Production stack: `/ws` routing, TLS, datastore auth (#23)
- Web resilience: no white screens, surfaced errors, WebSocket backoff (#24)
- Enforce the SSRF guard and load caps on the checker's bus path (#25)
- Gateway rate and body limits, server timeouts, API-key log redaction (#26)
- TTL retention on check and alert history (#37)

### Changed

- Bump `docker/login-action` from 3 to 4 (#28)
- Bump `docker/build-push-action` from 6 to 7 (#29)
- Bump `docker/setup-buildx-action` from 3 to 4 (#30)
- Bump `github.com/mattn/go-sqlite3` from 1.14.34 to 1.14.50 (#31)
- Bump `actions/checkout` from 4 to 7 (#32)
- Bump `github.com/rs/zerolog` from 1.34.0 to 1.35.1 (#33)
- Finish the remaining Dependabot bumps: progressbar, setup-node (#38)

## 2026-08-24: platform build

### Added

- Monorepo with an Echo gateway and NATS services, phases 0 to 2 (#2)
- MongoDB check history and Redis status and dedup (#3)
- React and TypeScript dashboard (#4)
- AI incident analysis with Claude, plus an MCP server (#5)
- CI/CD to GHCR, Kubernetes manifests, production compose, observability (#6)
- README, architecture diagram, docs and coverage polish (#7)
- App icon (#8)
- Testcontainers integration tests and a CI job for them (#12)
- Configurable CORS origins in the gateway (#13)
- Live status over WebSocket (#14)
- Monitors API and dynamic scheduler targets (#17)
- Multi-page observability dashboard with a Tailwind redesign (#18)

### Changed

- Milliseconds in API wire DTOs (#16)

### Fixed

- Security hardening: API-key auth, SSRF guard, load caps, RPC error surfacing (#9)
- Dependency-aware readiness probes and datastore timeouts (#10)
- Idempotent history writes on JetStream redelivery (#11)
- Secure the WebSocket with the API-key gate and Origin validation (#15)
- `urlguard` bare-host handling (#16)
