# rastime

Time utilities for safe timezone handling and time-of-day representations.

## Features

- `TimeOfDay` type for wall-clock times without dates
- Timezone-aware time creation and conversion
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
