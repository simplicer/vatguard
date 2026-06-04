# AGENTS.md

Architecture rules for the VAT Validator project. Agents and developers must follow these rules.

## Architecture

- **DDD**: Domain-Driven Design with vertical slicing and Screaming Architecture.
- **Bounded Context**: `vat-validation` — the sole context in this project.
- **Slice Location**: `internal/vatvalidation/`

## Layer Separation

### Domain (`internal/vatvalidation/domain/`)
- Pure business logic only.
- No framework dependencies.
- No infrastructure imports.
- Contains: entities, value objects, domain services, format rules, country catalog.
- Tests: colocated `*_test.go` files.

### Application (`internal/vatvalidation/application/`)
- Use cases only. One use case = one file.
- No persistence logic.
- No framework logic.
- No HTTP adapters.
- Contains: `ValidateVAT` use case and its interfaces/ports.
- Tests: colocated `*_test.go` files.

### Infrastructure (`internal/vatvalidation/infrastructure/`)
- HTTP handlers, routers, middleware.
- External API clients (VIES SOAP).
- Adapters that implement application ports.
- Contains: `http_handler.go`, `vies_client.go`, `router.go`.
- Tests: colocated `*_test.go` files.

## Dependency Direction

```
Infrastructure → Application → Domain
```

**NEVER** import infrastructure from domain or application.

## Naming Conventions

- Directories: `kebab-case`
- Files: `snake_case.go` (Go convention)
- Types/Interfaces: `PascalCase`
- Tests: `*_test.go` alongside source

## Testing Philosophy

- **Unit**: Every use case tested in isolation with mocked ports.
- **Integration**: VIES client, HTTP handler tested with `net/http/httptest`.
- **Acceptance**: Full validation flow (format + VIES) tested end-to-end.
- **E2E**: Full server behavior tested via HTTP requests.

Tests colocated with source (Go convention). No separate `tests/` directory.

## Logging

- Structured JSON logging via `log/slog`.
- Loki-compatible format.
- Logs stored under `var/log/`.

## Configuration

- Configuration files under `etc/`.
- `.env` at project root if needed by tooling.
- Environment variables for runtime config (PORT, LOG_LEVEL, etc.).

## Versioning

- Semantic Versioning: `X.Y.Z`
- Pre-release: `0.Y.Z`
- Version stored in `VERSION` file.
- Conventional Commits for all changes.

## External Dependencies

- Zero external dependencies. Go stdlib only.
- `net/http` for HTTP server.
- `log/slog` for structured logging.
- `encoding/json` for serialization.
- `net/http` for VIES SOAP client.
