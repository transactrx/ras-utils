package rastime

import (
	"encoding/json"
	"fmt"
	"time"
)

// Day of week consts
const (
	SUN = 0
	MON = 1
	TUE = 2
	WED = 3
	THU = 4
	FRI = 5
	SAT = 6
)

type TimeOfDay struct {
	Hour   int `json:"Hour"`
	Minute int `json:"Minute"`
}

// NewTimeOfDay creates a TimeOfDay from hour and minute
func NewTimeOfDay(hour, minute int) TimeOfDay {
	return TimeOfDay{Hour: hour, Minute: minute}
}

// TimeOfDayFromTime extracts hour and minute from time.Time
func TimeOfDayFromTime(t time.Time) TimeOfDay {
	return TimeOfDay{Hour: t.Hour(), Minute: t.Minute()}
}

// ParseTimeOfDay parses "HH:MM" string into TimeOfDay
func ParseTimeOfDay(s string) (TimeOfDay, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return TimeOfDay{}, fmt.Errorf("invalid time format %q: %w", s, err)
	}
	return TimeOfDay{Hour: t.Hour(), Minute: t.Minute()}, nil
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
