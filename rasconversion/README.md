# rasconversion

Type conversion helpers for PostgreSQL using pgx/pgtype. Converts nullable Go types to pgtype equivalents with proper null handling, and vice versa.

## Features

- Bidirectional conversion between Go types and pgtype
- Pointer-based API for nullable value handling
- Error-logging and error-returning variants
- OrDefault variants for value types with fallbacks
- JSONB marshaling/unmarshaling with generics

## Installation

```go
import "github.com/transactrx/ras-utils/rasconversion"
```

## Supported Types

| pgtype | Go Type | To pgtype | From pgtype | OrDefault |
|--------|---------|-----------|-------------|-----------|
| `Text` | `*string` | `ConvertToPgtypeString` | `ConvertFromPgtypeText` | `ConvertFromPgtypeTextOrDefault` |
| `Int2` | `*int32` / `*int16` | `ConvertToPgtypeInt2` | `ConvertFromPgtypeInt2` | `ConvertFromPgtypeInt2OrDefault` |
| `Int4` | `*int32` | `ConvertToPgtypeInt4` | `ConvertFromPgtypeInt4` | `ConvertFromPgtypeInt4OrDefault` |
| `Int8` | `*int64` | `ConvertToPgtypeInt8` | `ConvertFromPgtypeInt8` | `ConvertFromPgtypeInt8OrDefault` |
| `Bool` | `*bool` | `ConvertToPgtypeBool` | `ConvertFromPgtypeBool` | `ConvertFromPgtypeBoolOrDefault` |
| `Float4` | `*float32` | `ConvertToPgtypeFloat4` | `ConvertFromPgtypeFloat4` | `ConvertFromPgtypeFloat4OrDefault` |
| `Float8` | `*float64` | `ConvertToPgtypeFloat8` | `ConvertFromPgtypeFloat8` | `ConvertFromPgtypeFloat8OrDefault` |
| `Numeric` | `*float64` | `ConvertToPgtypeNumeric` | `ConvertFromPgtypeNumeric` | `ConvertFromPgtypeNumericOrDefault` |
| `Numeric` | `*big.Float` | — | `ConvertFromPgtypeNumericToBigFloat` | — |
| `Timestamp` | `*time.Time` | `ConvertToPgtypeTimestamp` | `ConvertFromPgtypeTimestamp` | `ConvertFromPgtypeTimestampOrDefault` |
| `Timestamptz` | `*time.Time` | `ConvertToPgtypeTimestamptz` | `ConvertFromPgtypeTimestamptz` | `ConvertFromPgtypeTimestamptzOrDefault` |
| `Date` | `*time.Time` | `ConvertToPgtypeDate` | `ConvertFromPgtypeDate` | `ConvertFromPgtypeDateOrDefault` |
| `Time` | `*time.Time` | `ConvertToPgtypeTime` | `ConvertFromPgtypeTime` | — |
| `Time` | `*rastime.TimeOfDay` | `ConvertTimeOfDayToPgtypeTime` | `ConvertPgtypeTimeToTimeOfDay` | — |
| `Interval` | `*time.Duration` | `ConvertToPgtypeInterval` | `ConvertFromPgtypeInterval` | `ConvertFromPgtypeIntervalOrDefault` |
| `UUID` | `*uuid.UUID` | `ConvertToPgtypeUUID` | `ConvertFromPgtypeUUID` | `ConvertFromPgtypeUUIDOrDefault` |
| `UUID` | `*string` | `ConvertToPgtypeUUIDFromString` | `ConvertFromPgtypeUUIDToString` | `ConvertFromPgtypeUUIDToStringOrDefault` |
| `JSONB` | `any` / `[]byte` | `ConvertToPgtypeJSONB` | `ConvertFromPgtypeJSONB[T]` | — |

## Usage

### To pgtype (Go → PostgreSQL)

```go
// Returns Valid: false if input pointer is nil
// Logs errors and returns Valid: false on conversion failure
pgText := rasconversion.ConvertToPgtypeString(stringPtr)
pgInt4 := rasconversion.ConvertToPgtypeInt4(int32Ptr)
pgFloat8 := rasconversion.ConvertToPgtypeFloat8(float64Ptr)
```

### Error-returning variants

```go
// Returns error instead of logging
pgText, err := rasconversion.TryConvertToPgtypeString(stringPtr)
pgNumeric, err := rasconversion.TryConvertToPgtypeNumeric(float64Ptr)
```

### From pgtype (PostgreSQL → Go)

```go
// Returns nil if pgtype has Valid: false
strPtr := rasconversion.ConvertFromPgtypeText(pgText)
int32Ptr := rasconversion.ConvertFromPgtypeInt4(pgInt4)
```

### OrDefault variants

```go
// Returns default value instead of nil when invalid
str := rasconversion.ConvertFromPgtypeTextOrDefault(pgText, "default")
num := rasconversion.ConvertFromPgtypeInt4OrDefault(pgInt4, 0)
```

### JSONB

```go
// Marshal any Go value to JSONB bytes
type UserPrefs struct {
    Theme    string `json:"theme"`
    Timezone string `json:"timezone"`
}
prefs := UserPrefs{Theme: "dark", Timezone: "America/New_York"}
jsonBytes := rasconversion.ConvertToPgtypeJSONB(prefs)

// Unmarshal JSONB bytes to a typed struct (generic)
result := rasconversion.ConvertFromPgtypeJSONB[UserPrefs](jsonBytes)
if result != nil {
    fmt.Println(result.Theme) // "dark"
}

// With error handling
result, err := rasconversion.TryConvertFromPgtypeJSONB[UserPrefs](jsonBytes)
```

### Interval

```go
// Duration to pgtype.Interval
d := 2*time.Hour + 30*time.Minute
pgInterval := rasconversion.ConvertToPgtypeInterval(&d)

// pgtype.Interval to Duration
duration := rasconversion.ConvertFromPgtypeInterval(pgInterval)

// With default fallback
duration := rasconversion.ConvertFromPgtypeIntervalOrDefault(pgInterval, time.Hour)
```
