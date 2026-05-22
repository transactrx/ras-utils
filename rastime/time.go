package rastime

import (
	"encoding/json"
	"fmt"
	"time"
)

// DayOfWeek represents a day of the week (0=Sunday, 6=Saturday)
type DayOfWeek int

const (
	SUN DayOfWeek = 0
	MON DayOfWeek = 1
	TUE DayOfWeek = 2
	WED DayOfWeek = 3
	THU DayOfWeek = 4
	FRI DayOfWeek = 5
	SAT DayOfWeek = 6
)

func (d DayOfWeek) String() string {
	names := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	if d < 0 || int(d) >= len(names) {
		return "Invalid"
	}
	return names[d]
}

// Short returns 3-letter abbreviation (Sun, Mon, etc.)
func (d DayOfWeek) Short() string {
	names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	if d < 0 || int(d) >= len(names) {
		return "???"
	}
	return names[d]
}

type TimeOfDay struct {
	Hour   int `json:"Hour"`
	Minute int `json:"Minute"`
}

// Validate returns an error if the TimeOfDay has invalid hour or minute values
func (t TimeOfDay) Validate() error {
	if t.Hour < 0 || t.Hour > 23 {
		return fmt.Errorf("invalid hour %d: must be 0-23", t.Hour)
	}
	if t.Minute < 0 || t.Minute > 59 {
		return fmt.Errorf("invalid minute %d: must be 0-59", t.Minute)
	}
	return nil
}

// NewTimeOfDay creates a TimeOfDay from hour and minute with validation
func NewTimeOfDay(hour, minute int) (TimeOfDay, error) {
	t := TimeOfDay{Hour: hour, Minute: minute}
	if err := t.Validate(); err != nil {
		return TimeOfDay{}, err
	}
	return t, nil
}

// TimeOfDayFromTime extracts hour and minute from time.Time
func TimeOfDayFromTime(t time.Time) TimeOfDay {
	return TimeOfDay{Hour: t.Hour(), Minute: t.Minute()}
}

// ParseTimeOfDay parses "HH:MM" (24-hour) or "H:MM PM" (12-hour) string into TimeOfDay
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

// String returns "HH:MM" format
func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour, t.Minute)
}

// ToMinutes converts to minutes since midnight for comparison
func (t TimeOfDay) ToMinutes() int {
	return t.Hour*60 + t.Minute
}

// Before returns true if t is before other
func (t TimeOfDay) Before(other TimeOfDay) bool {
	return t.ToMinutes() < other.ToMinutes()
}

// After returns true if t is after other
func (t TimeOfDay) After(other TimeOfDay) bool {
	return t.ToMinutes() > other.ToMinutes()
}

// Equal returns true if t equals other
func (t TimeOfDay) Equal(other TimeOfDay) bool {
	return t.ToMinutes() == other.ToMinutes()
}

// Between returns true if t is >= start and < end
func (t TimeOfDay) Between(start, end TimeOfDay) bool {
	return !t.Before(start) && t.Before(end)
}

// AddMinutes adds minutes to the time, wrapping at 24 hours
func (t TimeOfDay) AddMinutes(minutes int) TimeOfDay {
	totalMinutes := t.ToMinutes() + minutes
	// Wrap at 24 hours
	totalMinutes = totalMinutes % (24 * 60)
	if totalMinutes < 0 {
		totalMinutes += 24 * 60
	}
	return TimeOfDay{Hour: totalMinutes / 60, Minute: totalMinutes % 60}
}

// RoundUpToNextSlot rounds the time up to the next n-minute slot.
func (t TimeOfDay) RoundUpToNextSlot(slotMinutes int) TimeOfDay {
	remainder := t.Minute % slotMinutes
	if remainder == 0 {
		return t
	}
	return t.AddMinutes(slotMinutes - remainder)
}

// MarshalJSON writes TimeOfDay as "HH:MM" string
func (t TimeOfDay) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON reads TimeOfDay from either "HH:MM" string or {"Hour": X, "Minute": Y} object
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

// TimeRange represents an operating window within a day
type TimeRange struct {
	Start TimeOfDay `json:"start"`
	End   TimeOfDay `json:"end"`
}

// Contains returns true if the given time falls within this range
func (tr TimeRange) Contains(t TimeOfDay) bool {
	return !t.Before(tr.Start) && t.Before(tr.End)
}

// Duration returns the length of this time range
func (tr TimeRange) Duration() time.Duration {
	minutes := tr.End.ToMinutes() - tr.Start.ToMinutes()
	if minutes < 0 {
		minutes += 24 * 60 // handle overnight ranges
	}
	return time.Duration(minutes) * time.Minute
}

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
