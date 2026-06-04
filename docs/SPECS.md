# SPECS — VAT Validator API Business Logic

## 1. Domain: Country Classification

Countries are classified into four categories:

| Classification        | Description                                        | Examples                  |
|-----------------------|----------------------------------------------------|---------------------------|
| `EU_MEMBER`           | Current EU member states (27)                      | ES, DE, FR, IT, NL, ...   |
| `NON_EU_RECOGNIZED`   | Non-EU countries with known VAT formats            | GB, NO, CH, AU, JP, TR    |
| `NON_EU_UNKNOWN`      | Countries without recognized VAT format rules      | Various                   |
| `SANCTIONED`          | Countries under EU sanctions (prohibited)          | RU, BY                    |

### EU Member States (27)

| Country            | ISO Code | VIES Code | Format                                | Length   |
|--------------------|----------|-----------|---------------------------------------|----------|
| Austria            | AT       | AT        | U + 8 digits                          | 9        |
| Belgium            | BE       | BE        | 0 or 1 + 9 digits                     | 10       |
| Bulgaria           | BG       | BG        | 4-10 digits or 4-10 chars             | 4-10     |
| Croatia            | HR       | HR        | 11 digits                             | 11       |
| Cyprus             | CY       | CY        | 8 digits + 1 letter                   | 9        |
| Czech Republic     | CZ       | CZ        | 8-10 digits                           | 8-10     |
| Denmark            | DK       | DK        | 8 digits                              | 8        |
| Estonia            | EE       | EE        | 9 digits                              | 9        |
| Finland            | FI       | FI        | 8 digits                              | 8        |
| France             | FR       | FR        | 2 chars (alphanumeric) + 9 digits     | 11       |
| Germany            | DE       | DE        | 9 digits                              | 9        |
| Greece             | GR       | EL        | 9 digits                              | 9        |
| Hungary            | HU       | HU        | 8 digits                              | 8        |
| Ireland            | IE       | IE        | 7 digits + 1-2 letters (complex)      | 8-9      |
| Italy              | IT       | IT        | 11 digits                             | 11       |
| Latvia             | LV       | LV        | 11 digits                             | 11       |
| Lithuania          | LT       | LT        | 9 or 12 digits                        | 9 or 12  |
| Luxembourg         | LU       | LU        | 8 digits                              | 8        |
| Malta              | MT       | MT        | 6 digits                              | 8 (MT+6) |
| Netherlands        | NL       | NL        | 9 digits + B + 2 digits               | 12       |
| Poland             | PL       | PL        | 10 digits                             | 10       |
| Portugal           | PT       | PT        | 9 digits                              | 9        |
| Romania            | RO       | RO        | 2-10 digits                           | 2-10     |
| Slovakia           | SK       | SK        | 10 digits (must be divisible by 11)   | 10       |
| Slovenia           | SI       | SI        | 8 digits                              | 8        |
| Spain              | ES       | ES        | Letter + 7 chars + Letter/Digit       | 9        |
| Sweden             | SE       | SE        | 12 digits (last 2 = 01-12 or 99)      | 12       |

**Note**: Greece uses `EL` as its VIES country code (not `GR`).

### Sanctioned Countries

| Country    | ISO Code | Reason                       |
|------------|----------|------------------------------|
| Russia     | RU       | EU sanctions                 |
| Belarus    | BY       | EU sanctions                 |

## 2. Domain: VAT Number Parsing

A VAT number input is parsed as:

1. Extract the first 2 characters as potential country code (alphabetic).
2. The remainder is the national VAT number.
3. If the first 2 characters are not a recognized country code → `INVALID_FORMAT`.
4. If the country is `SANCTIONED` → reject with appropriate error.
5. Validate the national number against the country-specific format rules.

### Spain (ES) Specific Rules

- First character: letter (A-Z) but not N, S, or certain reserved letters.
- Last character: letter or digit.
- Middle: 7 alphanumeric characters.
- Common patterns: NIF (empresa), NIE (extranjero), CIF (legacy).

### Ireland (IE) Specific Rules

- Old format: 7 digits + 1 letter (A-W)
- New format: 7 digits + 2 letters
- Both accepted.

### Slovakia (SK) Specific Rules

- 10-digit number must be divisible by 11.

## 3. Application: Validate VAT Use Case

### Input

- `vat_number` (string): Full VAT number with country prefix (e.g., `ESA1234567J`).

### Output

- `country_code`: Extracted 2-letter country code.
- `country_classification`: `EU_MEMBER`, `NON_EU_RECOGNIZED`, `NON_EU_UNKNOWN`, or `SANCTIONED`.
- `format_valid`: Boolean indicating format compliance.
- `vies_registered`: Boolean (only for EU members, `null` otherwise).
- `vies_company_name`: String (only if VIES lookup succeeds).
- `vies_address`: String (only if VIES lookup succeeds).

### Flow

1. Parse VAT number → extract country code + national number.
2. Classify the country.
3. If `SANCTIONED` → return error with 422.
4. Validate format against country-specific rules.
5. If `EU_MEMBER` and format valid → query VIES.
6. Return combined result.

### VIES Integration

- Only for EU members with valid format.
- SOAP request to `https://ec.europa.eu/taxation_customs/vies/services/checkVatService`.
- 5-second timeout.
- If VIES is unavailable → return result with `vies_registered: null` and 503 status.

## 4. API Specification

### Endpoint

Two endpoints, one for format validation only and one for format + VIES:

```
GET /v1/vat/{number}   — Format validation only (fast, no external calls)
GET /v1/vies/{number}  — Format validation + VIES verification (EU only)
```

### Path Parameter

- `{number}`: VAT number including country prefix (e.g., `ES00000000T`, `DE123456789`).

### `/v1/vat/{number}` Behavior

1. Parse VAT number.
2. Classify country.
3. Validate format.
4. Return result (no VIES call).

### `/v1/vies/{number}` Behavior

1. Parse VAT number.
2. Classify country.
3. Validate format.
4. If EU member with valid format → query VIES.
5. Return combined result.

### Content-Type

- Success: `application/json`
- Error: `application/problem+json`

## 5. Non-Functional Requirements

- Startup time < 100ms
- Format validation < 50ms
- VIES lookup < 2s (including network)
- Binary size < 10MB
- Zero external dependencies
- Structured JSON logging (Loki-compatible)
