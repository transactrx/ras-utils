// Package rastime provides time-of-day utilities for schedule management.
//
// It includes [TimeOfDay] for representing hour/minute without a date,
// [TimeRange] for start/end operating windows, and [DayOfWeek] constants
// with validation, comparison, arithmetic, and JSON serialization support.
//
// TimeOfDay values serialize to JSON as "HH:MM" strings and can be parsed
// from both 24-hour ("14:30") and 12-hour ("2:30 PM") formats.
package rastime

import (
	"encoding/json"
	"fmt"
	"time"
)

// DayOfWeek represents a day of the week (0=Sunday, 6=Saturday)
type DayOfWeek int

// Day of week constants matching time.Weekday values.
const (
	SUN DayOfWeek = 0 // Sunday
	MON DayOfWeek = 1 // Monday
	TUE DayOfWeek = 2 // Tuesday
	WED DayOfWeek = 3 // Wednesday
	THU DayOfWeek = 4 // Thursday
	FRI DayOfWeek = 5 // Friday
	SAT DayOfWeek = 6 // Saturday
)

// String returns the full day name (e.g., "Monday").
func (d DayOfWeek) String() string {
	names := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	if d < 0 || int(d) >= len(names) {
		return "Invalid"
	}
	return names[d]
}

// Short returns the 3-letter abbreviation (e.g., "Mon").
func (d DayOfWeek) Short() string {
	names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	if d < 0 || int(d) >= len(names) {
		return "???"
	}
	return names[d]
}

// TimeOfDay represents a time of day as hour and minute, without a date component.
// It serializes to JSON as an "HH:MM" string.
type TimeOfDay struct {
	Hour   int `json:"Hour"`
	Minute int `json:"Minute"`
}

// DateRange represents a time period with a start and end timestamp.
// Used for period-based calculations like rolling annual windows.
type DateRange struct {
	Start time.Time `json:"Start"`
	End   time.Time `json:"End"`
}

// NewDateRange creates a DateRange with validation (Start must be before End).
func NewDateRange(start, end time.Time) (DateRange, error) {
	dr := DateRange{Start: start, End: end}
	if err := dr.Validate(); err != nil {
		return DateRange{}, err
	}
	return dr, nil
}

// CalendarYear returns a DateRange for the given year (Jan 1 00:00:00 to Jan 1 00:00:00 next year).
func CalendarYear(year int) DateRange {
	return DateRange{
		Start: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// AnnualPeriodFrom returns a DateRange starting at the given time and ending 1 year later.
func AnnualPeriodFrom(start time.Time) DateRange {
	return DateRange{
		Start: start,
		End:   start.AddDate(1, 0, 0),
	}
}

// Validate returns an error if Start is not before End.
func (dr DateRange) Validate() error {
	if !dr.Start.Before(dr.End) {
		return fmt.Errorf("invalid date range: start (%v) must be before end (%v)", dr.Start, dr.End)
	}
	return nil
}

// IsZero reports whether both Start and End are zero values.
func (dr DateRange) IsZero() bool {
	return dr.Start.IsZero() && dr.End.IsZero()
}

// Duration returns the length of this date range.
func (dr DateRange) Duration() time.Duration {
	return dr.End.Sub(dr.Start)
}

// Contains reports whether t is in the half-open interval [Start, End).
func (dr DateRange) Contains(t time.Time) bool {
	return !t.Before(dr.Start) && t.Before(dr.End)
}

// ContainsInclusive reports whether t is in the closed interval [Start, End].
func (dr DateRange) ContainsInclusive(t time.Time) bool {
	return !t.Before(dr.Start) && !t.After(dr.End)
}

// Overlaps reports whether this range overlaps with other.
// Adjacent ranges (where one ends exactly when the other starts) are not considered overlapping.
func (dr DateRange) Overlaps(other DateRange) bool {
	return dr.Start.Before(other.End) && dr.End.After(other.Start)
}

// NextAnnualPeriod returns a new DateRange shifted forward by 1 year.
func (dr DateRange) NextAnnualPeriod() DateRange {
	return DateRange{
		Start: dr.Start.AddDate(1, 0, 0),
		End:   dr.End.AddDate(1, 0, 0),
	}
}

// Validate returns an error if the TimeOfDay has invalid hour or minute values.
func (t TimeOfDay) Validate() error {
	if t.Hour < 0 || t.Hour > 23 {
		return fmt.Errorf("invalid hour %d: must be 0-23", t.Hour)
	}
	if t.Minute < 0 || t.Minute > 59 {
		return fmt.Errorf("invalid minute %d: must be 0-59", t.Minute)
	}
	return nil
}

// NewTimeOfDay creates a TimeOfDay from hour and minute with validation.
func NewTimeOfDay(hour, minute int) (TimeOfDay, error) {
	t := TimeOfDay{Hour: hour, Minute: minute}
	if err := t.Validate(); err != nil {
		return TimeOfDay{}, err
	}
	return t, nil
}

// TimeOfDayFromTime extracts hour and minute from a [time.Time].
func TimeOfDayFromTime(t time.Time) TimeOfDay {
	return TimeOfDay{Hour: t.Hour(), Minute: t.Minute()}
}

// ParseTimeOfDay parses a time string into a TimeOfDay.
// It accepts 24-hour format ("14:30") or 12-hour format ("2:30 PM", "2:30pm").
func ParseTimeOfDay(s string) (TimeOfDay, error) {
	// Try 24-hour format first
	if t, err := time.Parse("15:04", s); err == nil {
		return TimeOfDay{Hour: t.Hour(), Minute: t.Minute()}, nil
	}
	// Try 12-hour formats
	for _, layout := range []string{"3:04 PM", "3:04PM", "3:04 pm", "3:04pm"} {
		if t, err := time.Parse(layout, s); err == nil {
			return TimeOfDay{Hour: t.Hour(), Minute: t.Minute()}, nil
		}
	}
	return TimeOfDay{}, fmt.Errorf("invalid time format %q: expected HH:MM or H:MM AM/PM", s)
}

// String returns the time in "HH:MM" format.
func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour, t.Minute)
}

// ToMinutes converts to minutes since midnight for comparison.
func (t TimeOfDay) ToMinutes() int {
	return t.Hour*60 + t.Minute
}

// Before reports whether t is before other.
func (t TimeOfDay) Before(other TimeOfDay) bool {
	return t.ToMinutes() < other.ToMinutes()
}

// After reports whether t is after other.
func (t TimeOfDay) After(other TimeOfDay) bool {
	return t.ToMinutes() > other.ToMinutes()
}

// Equal reports whether t and other represent the same time.
func (t TimeOfDay) Equal(other TimeOfDay) bool {
	return t.ToMinutes() == other.ToMinutes()
}

// Between reports whether t is in the half-open interval [start, end).
func (t TimeOfDay) Between(start, end TimeOfDay) bool {
	return !t.Before(start) && t.Before(end)
}

// AddMinutes returns the time t+minutes, wrapping at midnight.
func (t TimeOfDay) AddMinutes(minutes int) TimeOfDay {
	totalMinutes := t.ToMinutes() + minutes
	// Wrap at 24 hours
	totalMinutes = totalMinutes % (24 * 60)
	if totalMinutes < 0 {
		totalMinutes += 24 * 60
	}
	return TimeOfDay{Hour: totalMinutes / 60, Minute: totalMinutes % 60}
}

// RoundUpToNextSlot rounds up to the next slotMinutes boundary.
// For example, 09:07 with slotMinutes=15 returns 09:15.
func (t TimeOfDay) RoundUpToNextSlot(slotMinutes int) TimeOfDay {
	remainder := t.Minute % slotMinutes
	if remainder == 0 {
		return t
	}
	return t.AddMinutes(slotMinutes - remainder)
}

// MarshalJSON implements [json.Marshaler], encoding as an "HH:MM" string.
func (t TimeOfDay) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON implements [json.Unmarshaler].
// It accepts either an "HH:MM" string or an object {"Hour": X, "Minute": Y}.
func (t *TimeOfDay) UnmarshalJSON(data []byte) error {
	// Try string format first ("HH:MM")
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		parsed, err := ParseTimeOfDay(s)
		if err != nil {
			return err
		}
		*t = parsed
		return nil
	}

	// Try object format ({"Hour": X, "Minute": Y})
	var obj struct {
		Hour   int `json:"Hour"`
		Minute int `json:"Minute"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("invalid TimeOfDay format: expected string or object, got %s", string(data))
	}
	*t = TimeOfDay{Hour: obj.Hour, Minute: obj.Minute}
	return nil
}

// TimeRange represents an operating window within a day, defined by start and end times.
type TimeRange struct {
	Start TimeOfDay `json:"start"`
	End   TimeOfDay `json:"end"`
}

// Contains reports whether t is in the half-open interval [Start, End).
func (tr TimeRange) Contains(t TimeOfDay) bool {
	return !t.Before(tr.Start) && t.Before(tr.End)
}

// Duration returns the length of this time range as a [time.Duration].
func (tr TimeRange) Duration() time.Duration {
	minutes := tr.End.ToMinutes() - tr.Start.ToMinutes()
	if minutes < 0 {
		minutes += 24 * 60 // handle overnight ranges
	}
	return time.Duration(minutes) * time.Minute
}

// Overlaps reports whether this range overlaps with any range in the provided slice.
// Adjacent ranges (e.g., 08:00-10:00 and 10:00-12:00) are not considered overlapping.
func (tr TimeRange) Overlaps(trs []TimeRange) bool {
	for _, item := range trs {
		// Two ranges overlap if one starts before the other ends AND ends after the other starts
		// This correctly handles adjacent ranges (e.g., 8-10 and 10-12) as non-overlapping
		if tr.Start.Before(item.End) && tr.End.After(item.Start) {
			return true
		}
	}
	return false
}
