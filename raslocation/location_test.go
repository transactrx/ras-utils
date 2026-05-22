package raslocation

import (
	"testing"
	"time"

	"github.com/transactrx/ras-utils/rastime"
)

func makeTimeOfDay(h, m int) rastime.TimeOfDay {
	return rastime.TimeOfDay{Hour: h, Minute: m}
}

func makeTimeRange(startH, startM, endH, endM int) rastime.TimeRange {
	return rastime.TimeRange{
		Start: makeTimeOfDay(startH, startM),
		End:   makeTimeOfDay(endH, endM),
	}
}

func TestDayHours_Validate(t *testing.T) {
	t.Run("valid - no overlap", func(t *testing.T) {
		dh := DayHours{
			DayOfWeek: 1,
			IsOpen:    true,
			TimeRanges: []rastime.TimeRange{
				makeTimeRange(9, 0, 12, 0),
				makeTimeRange(13, 0, 17, 0),
			},
		}
		if err := dh.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid - overlapping ranges", func(t *testing.T) {
		dh := DayHours{
			DayOfWeek: 1,
			IsOpen:    true,
			TimeRanges: []rastime.TimeRange{
				makeTimeRange(9, 0, 13, 0),
				makeTimeRange(12, 0, 17, 0),
			},
		}
		if err := dh.Validate(); err == nil {
			t.Error("expected error for overlapping ranges")
		}
	})

	t.Run("valid - adjacent ranges", func(t *testing.T) {
		dh := DayHours{
			DayOfWeek: 1,
			IsOpen:    true,
			TimeRanges: []rastime.TimeRange{
				makeTimeRange(9, 0, 12, 0),
				makeTimeRange(12, 0, 17, 0),
			},
		}
		if err := dh.Validate(); err != nil {
			t.Errorf("unexpected error for adjacent ranges: %v", err)
		}
	})
}

func TestDayHours_ContainsTime(t *testing.T) {
	dh := DayHours{
		DayOfWeek: 1,
		IsOpen:    true,
		TimeRanges: []rastime.TimeRange{
			makeTimeRange(9, 0, 12, 0),
			makeTimeRange(13, 0, 17, 0),
		},
	}

	tests := []struct {
		time rastime.TimeOfDay
		want bool
	}{
		{makeTimeOfDay(9, 0), true},
		{makeTimeOfDay(10, 30), true},
		{makeTimeOfDay(12, 0), false},  // lunch break
		{makeTimeOfDay(12, 30), false}, // lunch break
		{makeTimeOfDay(13, 0), true},
		{makeTimeOfDay(16, 59), true},
		{makeTimeOfDay(17, 0), false},
		{makeTimeOfDay(8, 0), false},
	}

	for _, tt := range tests {
		if got := dh.ContainsTime(tt.time); got != tt.want {
			t.Errorf("ContainsTime(%v) = %v, want %v", tt.time, got, tt.want)
		}
	}

	t.Run("closed day returns false", func(t *testing.T) {
		closed := DayHours{
			DayOfWeek:  0,
			IsOpen:     false,
			TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 17, 0)},
		}
		if closed.ContainsTime(makeTimeOfDay(12, 0)) {
			t.Error("closed day should return false")
		}
	})
}

func TestLocationHours_Validate(t *testing.T) {
	t.Run("valid schedule", func(t *testing.T) {
		lh := NewDefaultLocationHours(true, makeTimeOfDay(9, 0), makeTimeOfDay(17, 0))
		if err := lh.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid - day has overlapping ranges", func(t *testing.T) {
		lh := LocationHours{
			{
				DayOfWeek: 1,
				IsOpen:    true,
				TimeRanges: []rastime.TimeRange{
					makeTimeRange(9, 0, 14, 0),
					makeTimeRange(12, 0, 17, 0),
				},
			},
		}
		if err := lh.Validate(); err == nil {
			t.Error("expected error for overlapping ranges")
		}
	})
}

func TestLocationHours_String(t *testing.T) {
	t.Run("open schedule", func(t *testing.T) {
		lh := LocationHours{
			{DayOfWeek: 1, IsOpen: true, TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 17, 0)}},
			{DayOfWeek: 2, IsOpen: true, TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 17, 0)}},
		}
		got := lh.String()
		if got == "Closed" {
			t.Error("expected schedule string, got Closed")
		}
	})

	t.Run("all closed", func(t *testing.T) {
		lh := LocationHours{
			{DayOfWeek: 1, IsOpen: false},
			{DayOfWeek: 2, IsOpen: false},
		}
		if got := lh.String(); got != "Closed" {
			t.Errorf("got %q, want Closed", got)
		}
	})
}

func TestLocationHours_IsOpenAt(t *testing.T) {
	lh := LocationHours{
		{DayOfWeek: 1, IsOpen: true, TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 17, 0)}},  // Monday
		{DayOfWeek: 0, IsOpen: false, TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 17, 0)}}, // Sunday closed
	}

	// Monday at 10am
	monday := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC) // Jan 15, 2024 is Monday
	if !lh.IsOpenAt(monday) {
		t.Error("expected open on Monday at 10am")
	}

	// Monday at 6pm (closed)
	mondayEvening := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)
	if lh.IsOpenAt(mondayEvening) {
		t.Error("expected closed on Monday at 6pm")
	}

	// Sunday (closed)
	sunday := time.Date(2024, 1, 14, 10, 0, 0, 0, time.UTC)
	if lh.IsOpenAt(sunday) {
		t.Error("expected closed on Sunday")
	}
}

func TestLocationHours_IsOpenAtZone(t *testing.T) {
	lh := LocationHours{
		{DayOfWeek: 1, IsOpen: true, TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 17, 0)}},
	}

	// UTC time that's 10am in New York (EST = UTC-5)
	utcTime := time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC) // 15:00 UTC = 10:00 EST

	if !lh.IsOpenAtZone(utcTime, "America/New_York") {
		t.Error("expected open at 10am New York time")
	}

	t.Run("invalid timezone", func(t *testing.T) {
		if lh.IsOpenAtZone(utcTime, "Invalid/Zone") {
			t.Error("expected false for invalid timezone")
		}
	})
}

func TestLocationHours_MinutesUntilClose(t *testing.T) {
	lh := LocationHours{
		{DayOfWeek: 1, IsOpen: true, TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 17, 0)}},
	}

	t.Run("during open hours", func(t *testing.T) {
		monday := time.Date(2024, 1, 15, 16, 30, 0, 0, time.UTC) // 4:30pm, 30 min to close
		if got := lh.MinutesUntilClose(monday); got != 30 {
			t.Errorf("got %d, want 30", got)
		}
	})

	t.Run("outside open hours", func(t *testing.T) {
		monday := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)
		if got := lh.MinutesUntilClose(monday); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("closed day", func(t *testing.T) {
		sunday := time.Date(2024, 1, 14, 12, 0, 0, 0, time.UTC)
		if got := lh.MinutesUntilClose(sunday); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})
}

func TestLocationHours_WeeklyOpenMinutes(t *testing.T) {
	lh := LocationHours{
		{DayOfWeek: 1, IsOpen: true, TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 17, 0)}},  // 8 hours
		{DayOfWeek: 2, IsOpen: true, TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 17, 0)}},  // 8 hours
		{DayOfWeek: 0, IsOpen: false, TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 17, 0)}}, // closed
	}

	want := 2 * 8 * 60 // 2 days * 8 hours * 60 minutes
	if got := lh.WeeklyOpenMinutes(); got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestLocationHours_GetDayHours(t *testing.T) {
	lh := LocationHours{
		{DayOfWeek: 1, IsOpen: true, TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 17, 0)}},
		{DayOfWeek: 1, IsOpen: false, TimeRanges: []rastime.TimeRange{}},
		{DayOfWeek: 0, IsOpen: false},
	}

	t.Run("found open", func(t *testing.T) {
		dh := lh.GetDayHours(1, true)
		if dh == nil {
			t.Fatal("expected to find Monday open hours")
		}
		if !dh.IsOpen {
			t.Error("expected IsOpen = true")
		}
	})

	t.Run("found closed", func(t *testing.T) {
		dh := lh.GetDayHours(1, false)
		if dh == nil {
			t.Fatal("expected to find Monday closed entry")
		}
		if dh.IsOpen {
			t.Error("expected IsOpen = false")
		}
	})

	t.Run("not found", func(t *testing.T) {
		dh := lh.GetDayHours(6, true) // Saturday
		if dh != nil {
			t.Error("expected nil for Saturday")
		}
	})
}

func TestLocationHours_GetOpenDays(t *testing.T) {
	lh := LocationHours{
		{DayOfWeek: 1, IsOpen: true},
		{DayOfWeek: 2, IsOpen: true},
		{DayOfWeek: 0, IsOpen: false},
		{DayOfWeek: 6, IsOpen: false},
	}

	open := lh.GetOpenDays()
	if len(open) != 2 {
		t.Errorf("got %d open days, want 2", len(open))
	}
	for _, dh := range open {
		if !dh.IsOpen {
			t.Errorf("GetOpenDays returned closed day: %d", dh.DayOfWeek)
		}
	}
}

func TestLocationHours_GetFirstOpenWindow(t *testing.T) {
	lh := LocationHours{
		{
			DayOfWeek: 1,
			IsOpen:    true,
			TimeRanges: []rastime.TimeRange{
				makeTimeRange(9, 0, 12, 0),
				makeTimeRange(13, 0, 17, 0),
			},
		},
		{DayOfWeek: 0, IsOpen: false},
	}

	t.Run("found", func(t *testing.T) {
		tr := lh.GetFirstOpenWindow(1)
		if tr == nil {
			t.Fatal("expected to find first window")
		}
		if tr.Start.Hour != 9 || tr.End.Hour != 12 {
			t.Errorf("got %v, want 9:00-12:00", tr)
		}
	})

	t.Run("closed day", func(t *testing.T) {
		tr := lh.GetFirstOpenWindow(0)
		if tr != nil {
			t.Error("expected nil for closed day")
		}
	})

	t.Run("missing day", func(t *testing.T) {
		tr := lh.GetFirstOpenWindow(6)
		if tr != nil {
			t.Error("expected nil for missing day")
		}
	})
}

func TestLocationHours_GetNextOpenWindow(t *testing.T) {
	lh := LocationHours{
		{DayOfWeek: 1, IsOpen: true, TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 17, 0)}},  // Monday
		{DayOfWeek: 3, IsOpen: true, TimeRanges: []rastime.TimeRange{makeTimeRange(10, 0, 16, 0)}}, // Wednesday
	}

	t.Run("currently open - returns from time", func(t *testing.T) {
		monday := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC) // Monday noon
		got := lh.GetNextOpenWindow(monday)
		if !got.Equal(monday) {
			t.Errorf("got %v, want %v", got, monday)
		}
	})

	t.Run("before open - returns window start", func(t *testing.T) {
		monday := time.Date(2024, 1, 15, 7, 0, 0, 0, time.UTC) // Monday 7am
		got := lh.GetNextOpenWindow(monday)
		want := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("after close - returns next day", func(t *testing.T) {
		monday := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC) // Monday 6pm
		got := lh.GetNextOpenWindow(monday)
		want := time.Date(2024, 1, 17, 10, 0, 0, 0, time.UTC) // Wednesday 10am
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("no open window within 7 days", func(t *testing.T) {
		empty := LocationHours{}
		monday := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
		got := empty.GetNextOpenWindow(monday)
		if !got.IsZero() {
			t.Errorf("expected zero time, got %v", got)
		}
	})
}

func TestLocationHours_Merge(t *testing.T) {
	base := LocationHours{
		{DayOfWeek: 1, IsOpen: true, TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 17, 0)}},
		{DayOfWeek: 1, IsOpen: false, TimeRanges: []rastime.TimeRange{}},
		{DayOfWeek: 2, IsOpen: true, TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 17, 0)}},
	}

	override := LocationHours{
		{DayOfWeek: 1, IsOpen: true, TimeRanges: []rastime.TimeRange{makeTimeRange(10, 0, 16, 0)}}, // override Monday open
		{DayOfWeek: 3, IsOpen: true, TimeRanges: []rastime.TimeRange{makeTimeRange(9, 0, 12, 0)}},  // new Wednesday
	}

	merged := base.Merge(override)

	t.Run("overrides matching day+isOpen", func(t *testing.T) {
		dh := merged.GetDayHours(1, true)
		if dh == nil {
			t.Fatal("expected Monday open hours")
		}
		if dh.TimeRanges[0].Start.Hour != 10 {
			t.Errorf("expected start hour 10, got %d", dh.TimeRanges[0].Start.Hour)
		}
	})

	t.Run("preserves non-matching entries", func(t *testing.T) {
		dh := merged.GetDayHours(1, false)
		if dh == nil {
			t.Fatal("expected Monday closed entry to be preserved")
		}
	})

	t.Run("adds new entries", func(t *testing.T) {
		dh := merged.GetDayHours(3, true)
		if dh == nil {
			t.Fatal("expected Wednesday to be added")
		}
	})

	t.Run("preserves untouched entries", func(t *testing.T) {
		dh := merged.GetDayHours(2, true)
		if dh == nil {
			t.Fatal("expected Tuesday to be preserved")
		}
	})
}

func TestNewDefaultLocationHours(t *testing.T) {
	start := makeTimeOfDay(9, 0)
	end := makeTimeOfDay(17, 0)

	lh := NewDefaultLocationHours(true, start, end)

	if len(lh) != 7 {
		t.Errorf("expected 7 days, got %d", len(lh))
	}

	for i, dh := range lh {
		if dh.DayOfWeek != i {
			t.Errorf("day %d has wrong DayOfWeek: %d", i, dh.DayOfWeek)
		}
		if !dh.IsOpen {
			t.Errorf("day %d should be open", i)
		}
		if len(dh.TimeRanges) != 1 {
			t.Errorf("day %d should have 1 time range", i)
		}
	}
}
