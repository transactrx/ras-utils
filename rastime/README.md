# rastime

Time utilities for safe timezone handling, time-of-day representations, and date range operations.

## Features

- `TimeOfDay` type for wall-clock times without dates
- `TimeRange` for operating windows within a day
- `DateRange` for date/time periods with rolling and calendar year support
- `DayOfWeek` constants with string formatting
- Minutes-from-midnight arithmetic
- Range checking and comparisons

## Installation

```go
import "github.com/transactrx/ras-utils/rastime"
```

## Usage

### TimeOfDay

```go
// Create from hour:minute
tod, err := rastime.NewTimeOfDay(14, 30)  // 2:30 PM
if err != nil {
    // hour must be 0-23, minute 0-59
}

// Create from minutes since midnight
tod = rastime.TimeOfDayFromMinutes(870)  // 14:30

// Parse from string
tod, err = rastime.ParseTimeOfDay("14:30")
```

### TimeOfDay Operations

```go
// Convert to minutes since midnight
minutes := tod.MinutesSinceMidnight()  // 870

// Formatted output
str := tod.String()  // "14:30"

// Comparisons
if tod.Before(other) { ... }
if tod.After(other) { ... }
if tod.Equal(other) { ... }

// Check if within range
if tod.Between(startTod, endTod) { ... }
```

### Timezone-Aware Time Creation

```go
// Create time in a specific timezone
t, err := rastime.TimeInZone(2024, 6, 15, 14, 30, 0, "America/New_York")

// Get current time in a timezone
now, err := rastime.NowInZone("America/Chicago")
```

### Time Extraction

```go
// Extract TimeOfDay from a time.Time
tod := rastime.TimeOfDayFromTime(time.Now())

// Extract in a specific timezone
tod, err := rastime.TimeOfDayInZone(time.Now().UTC(), "America/Los_Angeles")
```

## API Reference

### Types

- `TimeOfDay` - Represents a wall-clock time (hour + minute) without a date

### Constructors

- `NewTimeOfDay(hour, minute int) (TimeOfDay, error)` - Create with validation
- `TimeOfDayFromMinutes(minutes int) TimeOfDay` - Create from minutes since midnight
- `ParseTimeOfDay(s string) (TimeOfDay, error)` - Parse "HH:MM" format
- `TimeOfDayFromTime(t time.Time) TimeOfDay` - Extract from time.Time
- `TimeOfDayInZone(t time.Time, timezone string) (TimeOfDay, error)` - Extract in timezone

### TimeOfDay Methods

- `MinutesSinceMidnight() int` - Convert to minutes since midnight
- `String() string` - Format as "HH:MM"
- `Before(other TimeOfDay) bool` - Comparison
- `After(other TimeOfDay) bool` - Comparison
- `Equal(other TimeOfDay) bool` - Comparison
- `Between(start, end TimeOfDay) bool` - Range check

### Time Functions

- `TimeInZone(year, month, day, hour, min, sec int, timezone string) (time.Time, error)`
- `NowInZone(timezone string) (time.Time, error)`

### DateRange

Represents a time period with start and end timestamps. Useful for period-based calculations like rolling annual windows.

```go
// Create a calendar year range (Jan 1 to Jan 1 next year)
dr := rastime.CalendarYear(2025)

// Create a rolling year from a specific date
firstAlert := time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC)
dr := rastime.AnnualPeriodFrom(firstAlert)

// Create with validation
dr, err := rastime.NewDateRange(start, end)
```

### DateRange Methods

```go
// Check if a time is within the range
if dr.Contains(t) { ... }           // half-open [Start, End)
if dr.ContainsInclusive(t) { ... }  // closed [Start, End]

// Get duration
duration := dr.Duration()

// Check for overlap with another range
if dr.Overlaps(other) { ... }

// Get the next annual period (shift forward 1 year)
nextPeriod := dr.NextAnnualPeriod()

// Validation
if err := dr.Validate(); err != nil { ... }  // Start must be before End
if dr.IsZero() { ... }  // both times are zero
```

### Rolling Period Example

```go
// Scenario: Patient's first alert starts a rolling annual period
// Period expires after 1 year, next alert starts a new period

firstAlert := time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC)
period1 := rastime.AnnualPeriodFrom(firstAlert)
// period1: 2025-02-20 to 2026-02-20

// Check if a date is in the period
dec1 := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
period1.Contains(dec1)  // true

// Period expired, next alert starts new period
mar30 := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)
period1.Contains(mar30)  // false (period expired)

period2 := rastime.AnnualPeriodFrom(mar30)
// period2: 2026-03-30 to 2027-03-30
```
