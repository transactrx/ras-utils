# rasvalidation

Common validators for identifiers, contact info, and formatted strings.

## Features

- UUID validation
- Email format validation (RFC 5322)
- US phone number validation
- US ZIP code validation
- NPI (National Provider Identifier) validation with Luhn checksum
- Date string validation

## Installation

```go
import "github.com/transactrx/ras-utils/rasvalidation"
```

## Usage

### UUID Validation

```go
if rasvalidation.IsValidUUID("550e8400-e29b-41d4-a716-446655440000") {
    // valid UUID format, non-nil
}
```

### Email Validation

```go
if rasvalidation.IsValidEmail("user@example.com") {
    // valid email format per RFC 5322
}
```

### Phone Number

```go
if rasvalidation.IsValidUSPhone("(555) 123-4567") {
    // valid US phone number (10 digits, various formats accepted)
}
```

### ZIP Code

```go
if rasvalidation.IsValidUSZip("12345") {
    // valid 5-digit ZIP
}

if rasvalidation.IsValidUSZip("12345-6789") {
    // valid ZIP+4
}
```

### NPI Validation

```go
// Format only (10 digits)
if rasvalidation.IsValidNPIFormat("1234567893") {
    // valid format
}

// Checksum only (assumes valid format)
if rasvalidation.IsValidNPIChecksum("1234567893") {
    // valid Luhn checksum
}

// Format + Luhn checksum
if rasvalidation.IsValidNPI("1234567893") {
    // valid NPI
}
```

### Date Validation

```go
// Generic date validation with custom layout
if rasvalidation.IsValidDateString("2024-06-15", "2006-01-02") {
    // valid for given layout
}

// ISO 8601 (YYYY-MM-DD)
if rasvalidation.IsValidISO8601Date("2024-06-15") {
    // valid
}

// MM/DD/YYYY
if rasvalidation.IsValidMMDDYYYYDate("06/15/2024") {
    // valid
}
```

## Deprecated

The following functions are deprecated. Use `rasformat` instead:

- `MaskPhone` - use `rasformat.MaskPhone`
- `MaskEmail` - use `rasformat.MaskEmail`
