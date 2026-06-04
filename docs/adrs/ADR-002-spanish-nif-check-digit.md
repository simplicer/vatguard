# ADR-002: Spanish NIF/NIE Check Digit Validation

## Status

Accepted

## Context

Spanish VAT numbers (NIF) have three distinct formats, each with specific check digit algorithms defined by the Spanish Ministerio del Interior and Real Decreto 1065/2007:

1. **NIF (DNI)**: 8 digits + control letter (mod 23)
2. **NIE**: X/Y/Z + 7 digits + control letter (X→0, Y→1, Z→2, then mod 23)
3. **NIF jurídica**: Entity type letter + 7 digits + control (digit or letter depending on entity type)

The control digit/letter must be validated, not just the format structure.

## Decision

We implement full check digit validation for all three Spanish NIF types. A VAT number that passes structural format checks but has an incorrect check digit is rejected with `format_valid: false`.

## Rationale

- Official source: Ministerio del Interior, Real Decreto 1065/2007
- Prevents false positives from well-formed but invalid VAT numbers
- Industry standard: all Spanish tax software validates the check digit
- The mod-23 letter table (TRWAGMYFPDXBNJZSQVHLCKE) and the entity-type control algorithm are deterministic

## Consequences

- Some previously "valid" Spanish VAT numbers now correctly fail validation
- Entity type letters are restricted to: A-H, J, N, P-S, U-W (excluding I, L, O, T, etc.)
- NIE prefix is restricted to X, Y, Z only

## Date

2025-01-01
