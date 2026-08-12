# rasformat

String formatting and masking utilities for the Clinical+ ecosystem.

## PHI Masking

Safe for logging PHI without exposing sensitive data:

```go
rasformat.MaskPhone("5551234567")           // "***-***-4567"
rasformat.MaskEmail("john@example.com")     // "j***@example.com"
rasformat.MaskName("John")                  // "J***"
rasformat.MaskFullName("John", "Smith")     // "J*** S***"
rasformat.MaskDOB("1990-05-15")             // "****-**-15"
```

## Phone Formatting

```go
rasformat.NormalizePhone("(555) 123-4567")  // "5551234567"
rasformat.FormatPhoneParens("5551234567")   // "(555) 123-4567"
rasformat.FormatPhoneDots("5551234567")     // "555.123.4567"
rasformat.FormatPhoneDashes("5551234567")   // "555-123-4567"
```

## Date Formatting

```go
rasformat.NormalizeDateISO("03/15/2024")    // "2024-03-15", nil
rasformat.NormalizeDateISO("20240315")      // "2024-03-15", nil
rasformat.NormalizeDateISO("2024-03-15T10:30:00Z") // "2024-03-15", nil
```

Supported input formats: RFC3339, ISO date (YYYY-MM-DD), compact (YYYYMMDD), US (MM/DD/YYYY, M/D/YYYY).
