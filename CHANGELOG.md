# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.8.0] - 2025-06-04

### Changed
- Renamed project from vatify to vatguard
- `/health` endpoint now returns `{"data": {"status": "ok"}}` consistent with the API envelope pattern
- Health is determined by the HTTP status code (200 = healthy)
- Updated OpenAPI 3.1 specification to reflect health response and version

## [0.7.0] - 2025-01-01

### Changed
- Split into two endpoints: `GET /v1/vat/{number}` (format only) and `GET /v1/vies/{number}` (format + VIES)
- VIES lookup is now opt-in via the `/vies` path instead of always-on
- Removed old `/v1/vat-numbers/{number}:validate` endpoint

### Added
- ADR-003: Separate /vat and /vies endpoints
- OpenAPI 3.1 specification at `docs/openapi.yaml`
- Application test for VIES skip behavior

## [0.6.0] - 2025-01-01

### Added
- Docker support: multi-stage Dockerfile (Go builder → scratch), docker-compose.yml
- ADR-002: Spanish NIF/NIE check digit validation decision record
- Makefile: `docker-build`, `docker-up`, `docker-down` targets

## [0.5.0] - 2025-01-01

### Added
- Full Spanish NIF/NIE check-digit validation per Ministerio del Interior
  - NIF (DNI): 8 digits + control letter (mod 23, table TRWAGMYFPDXBNJZSQVHLCKE)
  - NIE: X/Y/Z prefix → 0/1/2 + 7 digits + control letter
  - NIF jurídica: entity type letter + 7 digits + control digit/letter
- Entity type validation (A-H, J, N, P-S, U-W)
- 22 sub-tests for Spain covering DNI, NIE, NIF jurídica, and invalid cases

### Changed
- Spanish VAT numbers now require correct check digit (previously format-only)

## [0.4.0] - 2025-01-01

### Added
- VIES SOAP client with 5s timeout
- HTTP handler with Problem JSON (RFC 7807) error responses
- Router with structured logging middleware
- 8 VIES integration tests with httptest mock server
- 8 HTTP handler tests

## [0.3.0] - 2025-01-01

### Added
- ValidateVAT use case: parse → classify → validate format → VIES lookup
- VIESChecker port interface (inverted dependency)
- ValidationError with Problem JSON compatible fields
- 11 unit tests with mocked VIES

## [0.2.0] - 2025-01-01

### Added
- VAT number parsing: country code extraction + national number
- Format validation for all 27 EU member states
- Non-EU format validation: GB, NO, CH
- Special rules: SK divisible by 11, SE last 2 digits, IE old/new formats
- 46 format validation tests

## [0.1.0] - 2025-01-01

### Added
- Project scaffolding: go.mod, Makefile, README.md, AGENTS.md
- Country catalog: 27 EU members, 14 non-EU recognized, 2 sanctioned
- Case-insensitive country code lookup
- GR→EL VIES country code mapping
- 9 country classification tests

[0.8.0]: https://codeberg.org/simplicer/vatguard/releases/tag/v0.8.0
[0.7.0]: https://codeberg.org/simplicer/vatguard/releases/tag/v0.7.0
[0.6.0]: https://codeberg.org/simplicer/vatguard/releases/tag/v0.6.0
[0.5.0]: https://codeberg.org/simplicer/vatguard/releases/tag/v0.5.0
[0.4.0]: https://codeberg.org/simplicer/vatguard/releases/tag/v0.4.0
[0.3.0]: https://codeberg.org/simplicer/vatguard/releases/tag/v0.3.0
[0.2.0]: https://codeberg.org/simplicer/vatguard/releases/tag/v0.2.0
[0.1.0]: https://codeberg.org/simplicer/vatguard/releases/tag/v0.1.0
