# rasvalidation

Common validators for identifiers, contact info, and formatted strings.

## Features

- UUID validation
- Email format validation
- Phone number normalization and validation
- ZIP code validation
- NPI (National Provider Identifier) validation with checksum
- Date string parsing

## Installation

```go
import "github.com/transactrx/ras-utils/rasvalidation"
```

## Usage

### UUID Validation

```go
if rasvalidation.IsValidUUID("550e8400-e29b-41d4-a716-446655440000") {
    // valid UUID format
}
```

### Email Validation

```go
if rasvalidation.IsValidEmail("user@example.com") {
    // valid email format
}
```

### Phone Number

```go
// Normalize phone number (strips non-digits)
normalized := rasvalidation.NormalizePhone("(555) 123-4567")  // "5551234567"

// Validate (checks for 10 digits after normalization)
if rasvalidation.IsValidPhone("555-123-4567") {
    // valid US phone number
}
```

### ZIP Code

```go
// 5-digit ZIP
if rasvalidation.IsValidZIP("12345") {
    // valid
}

// ZIP+4
if rasvalidation.IsValidZIP("12345-6789") {
    // valid
}
```

### NPI Validation

```go
// Validates format and Luhn checksum
if rasvalidation.IsValidNPI("1234567893") {
    // valid NPI
}
```

### Date Parsing

```go
// Parse common date formats
t, err := rasvalidation.ParseDate("2024-06-15")
t, err := rasvalidation.ParseDate("06/15/2024")
t, err := rasvalidation.ParseDate("June 15, 2024")
```

## API Reference

### Functions

- `IsValidUUID(s string) bool` - UUID format validation
- `IsValidEmail(s string) bool` - Email format validation
- `NormalizePhone(s string) string` - Strip non-digits from phone number
- `IsValidPhone(s string) bool` - US phone number validation (10 digits)
- `IsValidZIP(s string) bool` - US ZIP code validation (5 or 9 digit)
- `IsValidNPI(s string) bool` - NPI validation with Luhn checksum
- `ParseDate(s string) (time.Time, error)` - Parse common date formats
