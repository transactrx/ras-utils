package raslocation

import (
	"fmt"
	"time"

	"github.com/transactrx/ras-utils/rastime"
)

// DayHours represents operating hours for a single day.
//
// Used in two contexts:
//  1. location_hours.hours JSONB: Default operating hours for a location
//  2. location_campaign_send_configuration.send_hours JSONB: Campaign-specific overrides
//
// The DailyMax field is only meaningful in the campaign config context,
// where it specifies the per-day send limit for that campaign at that location.
type DayHours struct {
	DayOfWeek  int                 `json:"day_of_week"`         // 0=Sunday, 6=Saturday
	IsOpen     bool                `json:"is_open"`             // Whether this day allows sending
	DailyMax   *int                `json:"daily_max,omitempty"` // Per-day limit (nil = use fallback of 100)
	TimeRanges []rastime.TimeRange `json:"time_ranges"`         // Operating windows for this day
}

// Validate checks for configuration errors like overlapping time ranges
func (dh DayHours) Validate() error {
	for i, tr := range dh.TimeRanges {
		if tr.Overlaps(dh.TimeRanges[i+1:]) {
			return fmt.Errorf("day %d has overlapping time ranges", dh.DayOfWeek)
		}
	}
	return nil
}

// ContainsTime returns true if the given time falls within any open range for this day
func (dh DayHours) ContainsTime(t rastime.TimeOfDay) bool {
	if !dh.IsOpen {
		return false
	}
	for _, tr := range dh.TimeRanges {
		if tr.Contains(t) {
			return true
		}
	}
	return false
}

// LocationHours is the full week schedule (maps to JSONB column)
type LocationHours []DayHours

// Validate checks all days for configuration errors
func (lh LocationHours) Validate() error {
	for _, dh := range lh {
		if err := dh.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// String returns a human-readable summary of the schedule
func (lh LocationHours) String() string {
	var parts []string
	for _, dh := range lh {
		if !dh.IsOpen {
			continue
		}
		day := rastime.DayOfWeek(dh.DayOfWeek).Short()
		for _, tr := range dh.TimeRanges {
			parts = append(parts, fmt.Sprintf("%s %s-%s", day, tr.Start, tr.End))
		}
	}
	if len(parts) == 0 {
		return "Closed"
	}
	return fmt.Sprintf("%v", parts)
}

// IsOpenAt returns true if location is open at the given time
func (lh LocationHours) IsOpenAt(t time.Time) bool {
	dayOfWeek := int(t.Weekday())
	timeOfDay := rastime.TimeOfDayFromTime(t)

	for _, dh := range lh {
		if dh.DayOfWeek == dayOfWeek {
			return dh.ContainsTime(timeOfDay)
		}
	}
	return false
}

// IsOpenAtZone returns true if location is open at the given time in the specified timezone.
// Returns false if the timezone is invalid.
func (lh LocationHours) IsOpenAtZone(t time.Time, zone string) bool {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return false
	}
	return lh.IsOpenAt(t.In(loc))
}

// MinutesUntilClose returns minutes until the current open window closes.
// Returns 0 if not currently open.
func (lh LocationHours) MinutesUntilClose(t time.Time) int {
	dayOfWeek := int(t.Weekday())
	timeOfDay := rastime.TimeOfDayFromTime(t)

	dh := lh.GetDayHours(dayOfWeek, true)
	if dh == nil {
		return 0
	}

	for _, tr := range dh.TimeRanges {
		if tr.Contains(timeOfDay) {
			return tr.End.ToMinutes() - timeOfDay.ToMinutes()
		}
	}
	return 0
}

// WeeklyOpenMinutes returns total open minutes across all days
func (lh LocationHours) WeeklyOpenMinutes() int {
	total := 0
	for _, dh := range lh {
		if !dh.IsOpen {
			continue
		}
		for _, tr := range dh.TimeRanges {
			total += int(tr.Duration().Minutes())
		}
	}
	return total
}

// GetDayHours returns hours for a specific day of week and isOpen value, or nil if not found
func (lh LocationHours) GetDayHours(dayOfWeek int, isOpen bool) *DayHours {
	for i := range lh {
		if lh[i].DayOfWeek == dayOfWeek && lh[i].IsOpen == isOpen {
			return &lh[i]
		}
	}
	return nil
}

// GetOpenDays returns all days where IsOpen is true
func (lh LocationHours) GetOpenDays() LocationHours {
	var result LocationHours
	for _, dh := range lh {
		if dh.IsOpen {
			result = append(result, dh)
		}
	}
	return result
}

// GetFirstOpenWindow returns the first open time range for a given day, or nil if closed
func (lh LocationHours) GetFirstOpenWindow(dayOfWeek int) *rastime.TimeRange {
	dh := lh.GetDayHours(dayOfWeek, true)
	if dh == nil || len(dh.TimeRanges) == 0 {
		return nil
	}
	return &dh.TimeRanges[0]
}

// GetNextOpenWindow returns the next time the location opens at or after the given time.
// Returns zero time if no open window found within 7 days.
func (lh LocationHours) GetNextOpenWindow(from time.Time) time.Time {
	for dayOffset := range 7 {
		checkDate := from.AddDate(0, 0, dayOffset)
		dayOfWeek := int(checkDate.Weekday())

		dh := lh.GetDayHours(dayOfWeek, true)
		if dh == nil {
			continue
		}

		for _, tr := range dh.TimeRanges {
			windowStart := time.Date(
				checkDate.Year(), checkDate.Month(), checkDate.Day(),
				tr.Start.Hour, tr.Start.Minute, 0, 0,
				from.Location(),
			)

			if dayOffset == 0 {
				windowEnd := time.Date(
					checkDate.Year(), checkDate.Month(), checkDate.Day(),
					tr.End.Hour, tr.End.Minute, 0, 0,
					from.Location(),
				)
				// If we're currently in this window, return from time
				if !from.Before(windowStart) && from.Before(windowEnd) {
					return from
				}
				// If window already passed, skip
				if from.After(windowEnd) || from.Equal(windowEnd) {
					continue
				}
			}

			if windowStart.After(from) || windowStart.Equal(from) {
				return windowStart
			}
		}
	}

	return time.Time{}
}

// Merge returns effective hours by overlaying overrides onto base hours.
// Override entries replace base entries only when both DayOfWeek and IsOpen match.
func (lh LocationHours) Merge(override LocationHours) LocationHours {
	result := make(LocationHours, len(lh))
	copy(result, lh)

	for _, oh := range override {
		found := false
		for i := range result {
			if result[i].DayOfWeek == oh.DayOfWeek && result[i].IsOpen == oh.IsOpen {
				result[i] = oh
				found = true
				break
			}
		}
		if !found {
			result = append(result, oh)
		}
	}
	return result
}

func NewDefaultLocationHours(isOpen bool, start rastime.TimeOfDay, end rastime.TimeOfDay) LocationHours {
	return LocationHours{
		{DayOfWeek: 0, IsOpen: isOpen, TimeRanges: []rastime.TimeRange{{Start: start, End: end}}},
		{DayOfWeek: 1, IsOpen: isOpen, TimeRanges: []rastime.TimeRange{{Start: start, End: end}}},
		{DayOfWeek: 2, IsOpen: isOpen, TimeRanges: []rastime.TimeRange{{Start: start, End: end}}},
		{DayOfWeek: 3, IsOpen: isOpen, TimeRanges: []rastime.TimeRange{{Start: start, End: end}}},
		{DayOfWeek: 4, IsOpen: isOpen, TimeRanges: []rastime.TimeRange{{Start: start, End: end}}},
		{DayOfWeek: 5, IsOpen: isOpen, TimeRanges: []rastime.TimeRange{{Start: start, End: end}}},
		{DayOfWeek: 6, IsOpen: isOpen, TimeRanges: []rastime.TimeRange{{Start: start, End: end}}},
	}
}
