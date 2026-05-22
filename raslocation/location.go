package raslocation

import (
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

// GetDayHours returns hours for a specific day of week, or nil if not found
func (lh LocationHours) GetDayHours(dayOfWeek int, isOpen bool) *DayHours {
	for i := range lh {
		if lh[i].DayOfWeek == dayOfWeek && lh[i].IsOpen == isOpen {
			return &lh[i]
		}
	}
	return nil
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
