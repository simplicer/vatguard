# ADR-003: Separate /vat and /vies Endpoints

## Status

Accepted

## Context

The initial design used a single endpoint `/v1/vat-numbers/{number}:validate` that always performed VIES lookup for valid EU VAT numbers. This had two issues:

1. VIES is an external service with ~1-2s latency. Most callers only need format validation.
2. The URL was verbose (`vat-numbers`) and used a Google-style `:action` suffix.

## Decision

Split into two clean, single-word endpoints:

- `GET /v1/vat/{number}` — format validation only (default, no VIES)
- `GET /v1/vies/{number}` — format validation + VIES verification

## Rationale

- **Separation of concerns**: Format validation is local and fast; VIES is remote and slow.
- **Zalando guideline**: Prefer short, descriptive resource names. One word > two words.
- **Default to fast**: Most callers want format validation. VIES should be opt-in.
- **No query parameters**: The distinction is in the resource path, not a flag.

## Consequences

- `/v1/vat/{number}` always returns `vies_registered: null`
- `/v1/vies/{number}` may return 503 if VIES is unreachable
- Old endpoint `/v1/vat-numbers/{number}:validate` is removed (no backward compat needed, pre-release)

## Date

2025-01-01
