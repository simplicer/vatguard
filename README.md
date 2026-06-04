# VAT Validator API

A lightweight, secure API in Go that validates VAT (Value Added Tax) identification numbers.

## Features

- Validate the **format** of VAT numbers for all 27 EU member states and non-EU countries
- Check whether a VAT number is **registered in VIES** (EU's VAT Information Exchange System)
- Classify countries as **EU member**, **non-EU recognized**, **non-EU unknown**, or **sanctioned**
- Zero external dependencies — Go stdlib only
- RESTful API following Zalando guidelines

## Setup

### Prerequisites

- Go 1.23+

### Build & Run

```bash
make build
make run
```

The server starts on `:8080` by default. Set the `PORT` environment variable to change.

### Docker

```bash
make docker-build
make docker-up
make docker-down
```

### Quick Test

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/v1/vat-numbers/ESA1234567J:validate
```

## Architecture

This project follows **Domain-Driven Design (DDD)** with **Screaming Architecture** and vertical slicing.

```
vat-validator/
├── cmd/server/                    # Application entry point
├── internal/vatvalidation/        # VAT Validation bounded context (slice)
│   ├── domain/                    # Pure business logic (entities, value objects, rules)
│   ├── application/               # Use cases
│   └── infrastructure/            # HTTP handlers, VIES client, adapters
├── etc/                           # Configuration files
├── var/log/                       # Runtime logs
├── bin/                           # Executables
├── docs/                          # Documentation and ADRs
├── Makefile                       # Build orchestration
└── go.mod                         # Go module definition
```

### Dependency Direction

```
Infrastructure → Application → Domain
```

The domain layer has zero imports from infrastructure or application. Application has zero imports from infrastructure.

## Testing

```bash
make test           # All tests with race detection
make test-cover     # Tests with HTML coverage report
make test-short     # Quick tests without race detection
```

Tests are colocated with source files (Go convention): `*_test.go` alongside `*.go`.

## Versioning

This project follows [Semantic Versioning](https://semver.org/). See `VERSION` file.

- Pre-release: `0.Y.Z`
- Conventional Commits are used for all changes

## API Specification

### Validate a VAT Number

```
GET /v1/vat-numbers/{vat-number}:validate
```

#### Response (200 OK)

```json
{
  "data": {
    "vat_number": "ESA1234567J",
    "country_code": "ES",
    "country_classification": "EU_MEMBER",
    "format_valid": true,
    "vies_registered": true,
    "vies_company_name": "COMPANY SL",
    "vies_address": "CALLE MAYOR 1, MADRID"
  },
  "meta": {
    "validated_at": "2025-01-01T12:00:00Z"
  }
}
```

#### Error Response (422 Unprocessable Entity)

```json
{
  "type": "https://vat-validator.dev/errors/invalid-format",
  "title": "Invalid VAT number format",
  "status": 422,
  "detail": "The VAT number 'XX123' does not match the expected format for country 'XX'",
  "instance": "/v1/vat-numbers/XX123:validate"
}
```

#### Error Response (503 Service Unavailable)

```json
{
  "type": "https://vat-validator.dev/errors/vies-unavailable",
  "title": "VIES service unavailable",
  "status": 503,
  "detail": "Could not reach the VIES SOAP service to verify the VAT number",
  "instance": "/v1/vat-numbers/ESA1234567J:validate"
}
```
