## Summary

<!-- What changes and why. -->

## Linked issue

Closes #

## Checklist

- [ ] Branched from `trunk`; this PR covers one change
- [ ] Commits are single-scope Conventional Commits
- [ ] `make lint` passes (go vet + gofmt)
- [ ] `make test` passes
- [ ] `make build` builds every service
- [ ] Integration tests pass if storage or messaging changed: `go test -tags=integration ./...` (needs Docker)
- [ ] Web changes: `pnpm type-check` and `pnpm build` pass in `apps/web`
- [ ] Docs, `.env.example` and `apps/gateway/openapi.yaml` updated if behavior or config changed
