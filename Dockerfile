# ==============================================================================
# Sitemon — one parametrized Dockerfile for every Go service.
# Pick the service at build time:  --build-arg SVC=gateway  (or checker/…/cli)
# Targets:  dev  (toolchain + source mount, used by compose)
#           prod (tiny static-ish runtime image)
# ==============================================================================

# ---- base toolchain ---------------------------------------------------------
FROM golang:1.27-alpine AS base
RUN apk add --no-cache git build-base
WORKDIR /src
ENV CGO_ENABLED=1

# ---- dev: compose mounts the source over /src and runs `go run` -------------
FROM base AS dev
# Source arrives via a bind mount at runtime; nothing to copy here.
CMD ["go", "run", "./apps/gateway/cmd"]

# ---- build: compile one service --------------------------------------------
FROM base AS build
ARG SVC=gateway
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -o /out/app ./apps/${SVC}/cmd

# ---- prod: minimal runtime --------------------------------------------------
FROM alpine:3.20 AS prod
RUN apk add --no-cache ca-certificates wget
COPY --from=build /out/app /usr/local/bin/app
ENTRYPOINT ["/usr/local/bin/app"]
