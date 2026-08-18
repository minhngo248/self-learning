# Agent Notes (build-your-own-lb/go)

This directory is a small Go "build your own load balancer" project (currently an Application Load Balancer + round-robin).

## Fast commands (PowerShell)

- Format: `gofmt -w .`
- Tests: `go test ./...`
- Vet: `go vet ./...`
- Build: `go build ./cmd/lb`
- Run: `go run ./cmd/lb --port 8080 --type application --algo roundrobin`

### Windows sandbox gotcha: `GOCACHE`

In some restricted environments, the default Go build cache under the user profile can fail with "Access is denied".
If that happens, set `GOCACHE` to a writable temp folder for the session:

`$env:GOCACHE = Join-Path $env:TEMP 'codex-go-cache-build-your-own-lb'`

Then rerun `go test`/`go vet`/`go build`.

## Project map

- Entry point: `cmd/lb/main.go`
- Proxy: `internal/proxy/alb.go` (reverse proxy to chosen backend)
- Algorithm(s): `internal/algo/*` (e.g. `roundrobin.go`)
- Backends: `internal/backend/backend.go` (backend address + counters/state)
- Health checking: `internal/health/*`
- CLI config/logging: `internal/config/*`

## Agent "skills" / workflow guidance

When making changes, optimize for small, verifiable edits:

- Prefer `rg` for search and `go test ./...` for validation.
- Keep changes scoped to the feature under work; avoid opportunistic refactors.
- Keep concurrency changes explicit and testable (avoid "fire and forget" goroutines unless bounded).
- Run `gofmt` before finalizing.

## Health-checking invariants (important)

If you change health checking, ensure these remain true:

- Health checks actually run while the LB is serving traffic (no placement after a blocking `ListenAndServe`).
- Health outcomes affect routing (the LB algorithm and/or proxy must not send traffic to removed/unhealthy backends).
- Health checks must:
  - enforce a timeout (do not rely on `http.DefaultClient` defaults),
  - treat non-2xx as unhealthy (if that's the intended semantics),
  - always close the response body to avoid leaks.
- Backend health state should be thread-safe (multiple checks and request paths may run concurrently).
- "Retry" semantics should be defined (consecutive failures vs lifetime failures) and reset/decay logic should match the definition.
