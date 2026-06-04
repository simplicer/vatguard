# ADR-001: Go Standard Library HTTP Server

## Status

Accepted

## Context

We need an HTTP server for the VAT Validator API. The Go ecosystem offers many third-party routers and frameworks (chi, gin, echo, gorilla/mux), but the standard library `net/http` has significantly improved since Go 1.22, now supporting method-based routing and path parameters natively.

## Decision

We will use the Go standard library `net/http` exclusively. No external HTTP frameworks or routers.

## Rationale

- **Zero dependencies**: The project constraint is zero external dependencies.
- **Go 1.22+ routing**: `http.ServeMux` now supports `GET /path/{param}` patterns with method matching.
- **Simplicity**: Fewer moving parts, less supply-chain risk.
- **Performance**: `net/http` is well-optimized and sufficient for our expected load.
- **Maintainability**: No dependency version conflicts or deprecation risk.

## Consequences

- Route definitions use Go 1.22+ `ServeMux` patterns.
- Middleware is implemented as standard `func(http.Handler) http.Handler` chains.
- No path validation utilities from third-party libraries; we implement our own.

## Date

2025-01-01
