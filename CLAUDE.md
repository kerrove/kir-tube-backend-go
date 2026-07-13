# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project status

This is an **early-stage scaffold**. As of this writing only `go.mod` / `go.sum`
are populated; `cmd/`, `internal/`, `pkg/`, `migrations/`, `configs/` are empty,
and `Dockerfile`, `docker-compose.yaml`, and `.env` are present but empty. There
is no `main` package or application code yet, and no git repository has been
initialized. Expect to be creating structure, not just editing it.

Module path: `go/kir-tube` (Go 1.26.3). "kir-tube" is a video-platform ("tube")
backend.

## Intended stack (inferred from dependencies)

The dependencies in `go.mod` are all marked `// indirect` because nothing imports
them yet, but they signal the planned architecture. When adding code, prefer
these already-vendored choices over introducing alternatives:

- **Database:** PostgreSQL via `jackc/pgx/v5` (driver/pool) with **GORM**
  (`gorm.io/gorm` + `gorm.io/driver/postgres`) as the ORM. `migrations/` is the
  home for schema migrations.
- **Config:** `joho/godotenv` — load `.env` at startup. `configs/` holds config
  files/loaders. Do not commit real secrets to `.env` (it is currently empty).
- **Validation:** `go-playground/validator/v10` for request/struct validation.
- **Auth/crypto:** `golang.org/x/crypto` (e.g. bcrypt for password hashing).
- **File handling:** `gabriel-vasile/mimetype` for MIME sniffing (uploads).
- **Concurrency:** `golang.org/x/sync`.

No HTTP router/framework is pinned yet — choose one (net/http, chi, gin, echo)
when the HTTP layer is added, and confirm the direction if it isn't already
established in code.

## Layout convention

Standard Go project layout is implied by the directory names:

- `cmd/` — entrypoints; one subdirectory per binary, each with `package main`
  (e.g. `cmd/api/main.go`).
- `internal/` — private application code, not importable by other modules
  (handlers, services, repositories, domain models).
- `pkg/` — library code intended to be reusable/importable.
- `migrations/` — database migrations.
- `configs/` — configuration files and loading.

## Commands

```sh
# Build everything
go build ./...

# Run the API (once an entrypoint exists, e.g. cmd/api)
go run ./cmd/api

# Test
go test ./...                        # all packages
go test ./internal/...               # a subtree
go test -run TestName ./internal/foo # a single test
go test -race -cover ./...           # race detector + coverage

# Formatting and static checks
gofmt -w .
go vet ./...

# Dependency hygiene (indirect deps become direct as code imports them)
go mod tidy
```

Docker (`Dockerfile` / `docker-compose.yaml`) is scaffolded but empty; there is
no working container build yet.
