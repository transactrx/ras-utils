# raslocation

Location-specific time and scheduling utilities for the Clinical+ ecosystem.

## Features

- Operating hours management
- Timezone-aware open/close checks
- Schedule merging (base + campaign overrides)
- Validation for overlapping time ranges
- Weekly open minutes calculation

## Installation

```go
import "github.com/transactrx/ras-utils/raslocation"
import "github.com/transactrx/ras-utils/rastime"
```

## Usage

### Creating Location Hours

```go
// Create default hours (same hours all 7 days)
start, _ := rastime.NewTimeOfDay(9, 0)
end, _ := rastime.NewTimeOfDay(17, 0)
hours := raslocation.NewDefaultLocationHours(true, start, end)
```

### Checking Open Status

```go
// Check if open at a specific time
if hours.IsOpenAt(time.Now()) {
    // location is open
}

// Check with timezone conversion
if hours.IsOpenAtZone(time.Now().UTC(), "America/New_York") {
    // open in New York time
}
```

### Time Calculations

```go
// Get minutes until close (0 if not open)
minutes := hours.MinutesUntilClose(time.Now())

// Get total weekly open minutes
totalMinutes := hours.WeeklyOpenMinutes()

// Find next open window
nextOpen := hours.GetNextOpenWindow(time.Now())
if !nextOpen.IsZero() {
    fmt.Printf("Opens at %v\n", nextOpen)
}
```

### Schedule Merging

```go
// Merge campaign overrides onto base hours
// Only replaces entries where DayOfWeek AND IsOpen match
effective := baseHours.Merge(campaignOverrides)
```

### Validation

```go
// Validate for overlapping time ranges
if err := hours.Validate(); err != nil {
    log.Printf("Invalid schedule: %v", err)
}
```

### Querying

```go
// Get open days only
openDays := hours.GetOpenDays()

// Human-readable schedule
fmt.Println(hours.String()) // "[Mon 09:00-17:00 Tue 09:00-17:00 ...]"
```

## API Reference

### Types

- `LocationHours` - Collection of daily hour entries
- `LocationHourEntry` - Single day's operating hours

### Functions

- `NewDefaultLocationHours(isOpen bool, start, end rastime.TimeOfDay) LocationHours`

### Methods

- `IsOpenAt(t time.Time) bool`
- `IsOpenAtZone(t time.Time, timezone string) bool`
- `MinutesUntilClose(t time.Time) int`
- `WeeklyOpenMinutes() int`
- `GetNextOpenWindow(t time.Time) time.Time`
- `Merge(overrides LocationHours) LocationHours`
- `Validate() error`
- `GetOpenDays() LocationHours`
- `String() string`
