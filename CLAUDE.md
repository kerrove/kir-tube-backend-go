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

# Development Architecture Guidelines

This document outlines the core architectural principles and patterns required for this project. All code contributions must adhere to these standards to ensure maintainability, scalability, and testability.

---

## 1. Dependency Injection (DI)

We use Dependency Injection to achieve **Inversion of Control (IoC)**. Components must not instantiate their dependencies directly. Instead, dependencies must be provided (injected) from the outside.

### Guidelines

- **Constructor Injection:** Prefer constructor injection for all required dependencies. This makes dependencies explicit and simplifies testing.
- **Interface-Driven:** Program to interfaces, not implementations. Inject the interface type to allow swapping implementations (e.g., swapping a production database service for a mock repository in tests).
- **Decoupling:** High-level modules must not depend on low-level modules; both must depend on abstractions.

---

## 2. SOLID Principles

Every module, class, and function must respect the five SOLID principles of object-oriented design:

- **Single Responsibility Principle (SRP):** A class should have one, and only one, reason to change. Separate business logic, data access, and presentation layers.
- **Open/Closed Principle (OCP):** Software entities should be open for extension, but closed for modification. Use polymorphism and interfaces to add new behavior without altering existing code.
- **Liskov Substitution Principle (LSP):** Subtypes must be completely substitutable for their base types without breaking the application behavior.
- **Interface Segregation Principle (ISP):** Clients should not be forced to depend on methods they do not use. Prefer many small, client-specific interfaces over one large, general-purpose interface.
- **Dependency Inversion Principle (DIP):** Depend on abstractions, not concretions. (See the _Dependency Injection_ section above).

---

## 3. Domain-Driven Design (DDD)

The codebase is structured around the business domain. We separate the technical implementation details from the core business logic.

### Core Concepts to Follow

- **Bounded Contexts:** Clear boundaries must be defined around specific parts of the domain. Avoid bleeding models across different contexts (e.g., a `User` in an Auth context is different from a `Customer` in a Billing context).
- **Ubiquitous Language:** Code terminology (class names, methods, variables) must strictly match the business vocabulary used by domain experts.
- **Domain Models:**
  - **Entities:** Objects with a distinct identity that persists over time (e.g., `Order` with a unique ID).
  - **Value Objects:** Objects defined solely by their attributes, with no conceptual identity. They must be **immutable** (e.g., `Money`, `Address`).
  - **Aggregates:** A cluster of associated objects treated as a single unit for data changes. Every aggregate has a root entity through which all external interactions must pass.

---

## 4. Clean Architecture

The application enforces a strict separation of concerns using a layered approach, ensuring that the business logic is independent of frameworks, UI, and databases.

### The Dependency Rule

Source code dependencies must only point **inwards**, toward the core business logic. The inner layers must know nothing about the outer layers.

### Layer Hierarchy (Inner to Outer)

1. **Entities (Domain Layer):** \* Encapsulates enterprise-wide business rules and core domain models.
   - Has zero dependencies on external frameworks, libraries, or databases.
2. **Use Cases (Application Layer):** \* Contains application-specific business rules.
   - Orchestrates the flow of data to and from the entities.
   - Defines interfaces for external dependencies (e.g., `UserRepositoryInterface`).
3. **Interface Adapters (Presentation/Infrastructure Layer):** \* Converts data from the format most convenient for use cases and entities to the format most convenient for external agencies (UI, DB, Web APIs).
   - Includes Controllers, Presenters, and Repository implementations.
4. **Frameworks & Drivers:** \* The outermost layer consisting of tools such as databases, web frameworks, UI engines, and third-party SDKs.

---

## Code Review Checklist for AI & Contributors

- [ ] Are dependencies injected via constructors rather than hardcoded with `new`?
- [ ] Does this class have a single responsibility?
- [ ] Are domain models free of framework-specific annotations/dependencies where possible?
- [ ] Does the data flow respect the Clean Architecture dependency rule (no outer-layer leaks into the inner core)?
