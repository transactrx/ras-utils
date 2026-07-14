package rastime

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDayOfWeek_String(t *testing.T) {
	tests := []struct {
		day  DayOfWeek
		want string
	}{
		{SUN, "Sunday"},
		{MON, "Monday"},
		{TUE, "Tuesday"},
		{WED, "Wednesday"},
		{THU, "Thursday"},
		{FRI, "Friday"},
		{SAT, "Saturday"},
		{DayOfWeek(-1), "Invalid"},
		{DayOfWeek(7), "Invalid"},
	}

	for _, tt := range tests {
		if got := tt.day.String(); got != tt.want {
			t.Errorf("DayOfWeek(%d).String() = %q, want %q", tt.day, got, tt.want)
		}
	}
}

func TestDayOfWeek_Short(t *testing.T) {
	tests := []struct {
		day  DayOfWeek
		want string
	}{
		{SUN, "Sun"},
		{MON, "Mon"},
		{SAT, "Sat"},
		{DayOfWeek(-1), "???"},
		{DayOfWeek(7), "???"},
	}

	for _, tt := range tests {
		if got := tt.day.Short(); got != tt.want {
			t.Errorf("DayOfWeek(%d).Short() = %q, want %q", tt.day, got, tt.want)
		}
	}
}

func TestTimeOfDay_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tod     TimeOfDay
		wantErr bool
	}{
		{"valid midnight", TimeOfDay{0, 0}, false},
		{"valid noon", TimeOfDay{12, 30}, false},
		{"valid end of day", TimeOfDay{23, 59}, false},
		{"invalid hour negative", TimeOfDay{-1, 0}, true},
		{"invalid hour too high", TimeOfDay{24, 0}, true},
		{"invalid minute negative", TimeOfDay{12, -1}, true},
		{"invalid minute too high", TimeOfDay{12, 60}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tod.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewTimeOfDay(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tod, err := NewTimeOfDay(9, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tod.Hour != 9 || tod.Minute != 30 {
			t.Errorf("got %v, want {9, 30}", tod)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := NewTimeOfDay(25, 0)
		if err == nil {
			t.Error("expected error for invalid hour")
		}
	})
}

func TestTimeOfDayFromTime(t *testing.T) {
	tm := time.Date(2024, 1, 15, 14, 45, 30, 0, time.UTC)
	tod := TimeOfDayFromTime(tm)

	if tod.Hour != 14 || tod.Minute != 45 {
		t.Errorf("got %v, want {14, 45}", tod)
	}
}

func TestParseTimeOfDay(t *testing.T) {
	tests := []struct {
		input   string
		want    TimeOfDay
		wantErr bool
	}{
		{"09:30", TimeOfDay{9, 30}, false},
		{"14:00", TimeOfDay{14, 0}, false},
		{"00:00", TimeOfDay{0, 0}, false},
		{"23:59", TimeOfDay{23, 59}, false},
		{"9:30 AM", TimeOfDay{9, 30}, false},
		{"9:30 PM", TimeOfDay{21, 30}, false},
		{"12:00 PM", TimeOfDay{12, 0}, false},
		{"12:00 AM", TimeOfDay{0, 0}, false},
		{"9:30AM", TimeOfDay{9, 30}, false},
		{"9:30pm", TimeOfDay{21, 30}, false},
		{"invalid", TimeOfDay{}, true},
		{"25:00", TimeOfDay{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseTimeOfDay(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimeOfDay(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseTimeOfDay(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTimeOfDay_String(t *testing.T) {
	tests := []struct {
		tod  TimeOfDay
		want string
	}{
		{TimeOfDay{9, 5}, "09:05"},
		{TimeOfDay{14, 30}, "14:30"},
		{TimeOfDay{0, 0}, "00:00"},
	}

	for _, tt := range tests {
		if got := tt.tod.String(); got != tt.want {
			t.Errorf("TimeOfDay{%d, %d}.String() = %q, want %q", tt.tod.Hour, tt.tod.Minute, got, tt.want)
		}
	}
}

func TestTimeOfDay_ToMinutes(t *testing.T) {
	tests := []struct {
		tod  TimeOfDay
		want int
	}{
		{TimeOfDay{0, 0}, 0},
		{TimeOfDay{1, 0}, 60},
		{TimeOfDay{9, 30}, 570},
		{TimeOfDay{23, 59}, 1439},
	}

	for _, tt := range tests {
		if got := tt.tod.ToMinutes(); got != tt.want {
			t.Errorf("TimeOfDay{%d, %d}.ToMinutes() = %d, want %d", tt.tod.Hour, tt.tod.Minute, got, tt.want)
		}
	}
}

func TestTimeOfDay_Comparisons(t *testing.T) {
	early := TimeOfDay{9, 0}
	late := TimeOfDay{17, 0}
	same := TimeOfDay{9, 0}

	if !early.Before(late) {
		t.Error("expected 9:00 before 17:00")
	}
	if early.Before(same) {
		t.Error("9:00 should not be before 9:00")
	}

	if !late.After(early) {
		t.Error("expected 17:00 after 9:00")
	}
	if late.After(late) {
		t.Error("17:00 should not be after 17:00")
	}

	if !early.Equal(same) {
		t.Error("expected 9:00 equal to 9:00")
	}
	if early.Equal(late) {
		t.Error("9:00 should not equal 17:00")
	}
}

func TestTimeOfDay_Between(t *testing.T) {
	start := TimeOfDay{9, 0}
	end := TimeOfDay{17, 0}

	tests := []struct {
		tod  TimeOfDay
		want bool
	}{
		{TimeOfDay{9, 0}, true},   // inclusive start
		{TimeOfDay{12, 0}, true},  // middle
		{TimeOfDay{16, 59}, true}, // just before end
		{TimeOfDay{17, 0}, false}, // exclusive end
		{TimeOfDay{8, 59}, false}, // before start
		{TimeOfDay{17, 1}, false}, // after end
	}

	for _, tt := range tests {
		if got := tt.tod.Between(start, end); got != tt.want {
			t.Errorf("%v.Between(%v, %v) = %v, want %v", tt.tod, start, end, got, tt.want)
		}
	}
}

func TestTimeOfDay_AddMinutes(t *testing.T) {
	tests := []struct {
		tod     TimeOfDay
		minutes int
		want    TimeOfDay
	}{
		{TimeOfDay{9, 0}, 30, TimeOfDay{9, 30}},
		{TimeOfDay{9, 30}, 60, TimeOfDay{10, 30}},
		{TimeOfDay{23, 30}, 60, TimeOfDay{0, 30}},   // wrap forward
		{TimeOfDay{0, 30}, -60, TimeOfDay{23, 30}},  // wrap backward
		{TimeOfDay{12, 0}, 0, TimeOfDay{12, 0}},     // no change
		{TimeOfDay{0, 0}, 1440, TimeOfDay{0, 0}},    // full day
		{TimeOfDay{0, 0}, -1440, TimeOfDay{0, 0}},   // full day backward
	}

	for _, tt := range tests {
		got := tt.tod.AddMinutes(tt.minutes)
		if got != tt.want {
			t.Errorf("%v.AddMinutes(%d) = %v, want %v", tt.tod, tt.minutes, got, tt.want)
		}
	}
}

func TestTimeOfDay_RoundUpToNextSlot(t *testing.T) {
	tests := []struct {
		tod         TimeOfDay
		slotMinutes int
		want        TimeOfDay
	}{
		{TimeOfDay{9, 0}, 15, TimeOfDay{9, 0}},   // already on slot
		{TimeOfDay{9, 1}, 15, TimeOfDay{9, 15}},  // round up
		{TimeOfDay{9, 14}, 15, TimeOfDay{9, 15}}, // round up
		{TimeOfDay{9, 15}, 15, TimeOfDay{9, 15}}, // already on slot
		{TimeOfDay{9, 16}, 15, TimeOfDay{9, 30}}, // round up
		{TimeOfDay{9, 7}, 30, TimeOfDay{9, 30}},  // 30 min slots
	}

	for _, tt := range tests {
		got := tt.tod.RoundUpToNextSlot(tt.slotMinutes)
		if got != tt.want {
			t.Errorf("%v.RoundUpToNextSlot(%d) = %v, want %v", tt.tod, tt.slotMinutes, got, tt.want)
		}
	}
}

func TestTimeOfDay_JSON(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		tod := TimeOfDay{9, 30}
		data, err := json.Marshal(tod)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		if string(data) != `"09:30"` {
			t.Errorf("got %s, want %q", data, "09:30")
		}
	})

	t.Run("unmarshal string", func(t *testing.T) {
		var tod TimeOfDay
		if err := json.Unmarshal([]byte(`"14:30"`), &tod); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if tod.Hour != 14 || tod.Minute != 30 {
			t.Errorf("got %v, want {14, 30}", tod)
		}
	})

	t.Run("unmarshal object", func(t *testing.T) {
		var tod TimeOfDay
		if err := json.Unmarshal([]byte(`{"Hour": 9, "Minute": 15}`), &tod); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if tod.Hour != 9 || tod.Minute != 15 {
			t.Errorf("got %v, want {9, 15}", tod)
		}
	})

	t.Run("unmarshal invalid", func(t *testing.T) {
		var tod TimeOfDay
		if err := json.Unmarshal([]byte(`123`), &tod); err == nil {
			t.Error("expected error for invalid format")
		}
	})
}

func TestTimeRange_Contains(t *testing.T) {
	tr := TimeRange{
		Start: TimeOfDay{9, 0},
		End:   TimeOfDay{17, 0},
	}

	tests := []struct {
		tod  TimeOfDay
		want bool
	}{
		{TimeOfDay{9, 0}, true},
		{TimeOfDay{12, 0}, true},
		{TimeOfDay{16, 59}, true},
		{TimeOfDay{17, 0}, false},
		{TimeOfDay{8, 59}, false},
		{TimeOfDay{17, 1}, false},
	}

	for _, tt := range tests {
		if got := tr.Contains(tt.tod); got != tt.want {
			t.Errorf("TimeRange.Contains(%v) = %v, want %v", tt.tod, got, tt.want)
		}
	}
}

func TestTimeRange_Duration(t *testing.T) {
	tests := []struct {
		name string
		tr   TimeRange
		want time.Duration
	}{
		{
			"standard range",
			TimeRange{TimeOfDay{9, 0}, TimeOfDay{17, 0}},
			8 * time.Hour,
		},
		{
			"short range",
			TimeRange{TimeOfDay{9, 0}, TimeOfDay{9, 30}},
			30 * time.Minute,
		},
		{
			"same start and end",
			TimeRange{TimeOfDay{0, 0}, TimeOfDay{0, 0}},
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.Duration(); got != tt.want {
				t.Errorf("Duration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeRange_Overlaps(t *testing.T) {
	tests := []struct {
		name   string
		tr     TimeRange
		others []TimeRange
		want   bool
	}{
		{
			"no overlap - before",
			TimeRange{TimeOfDay{9, 0}, TimeOfDay{10, 0}},
			[]TimeRange{{TimeOfDay{10, 0}, TimeOfDay{11, 0}}},
			false,
		},
		{
			"no overlap - after",
			TimeRange{TimeOfDay{11, 0}, TimeOfDay{12, 0}},
			[]TimeRange{{TimeOfDay{9, 0}, TimeOfDay{10, 0}}},
			false,
		},
		{
			"overlap - partial",
			TimeRange{TimeOfDay{9, 0}, TimeOfDay{11, 0}},
			[]TimeRange{{TimeOfDay{10, 0}, TimeOfDay{12, 0}}},
			true,
		},
		{
			"overlap - contained",
			TimeRange{TimeOfDay{10, 0}, TimeOfDay{11, 0}},
			[]TimeRange{{TimeOfDay{9, 0}, TimeOfDay{12, 0}}},
			true,
		},
		{
			"overlap - contains",
			TimeRange{TimeOfDay{9, 0}, TimeOfDay{12, 0}},
			[]TimeRange{{TimeOfDay{10, 0}, TimeOfDay{11, 0}}},
			true,
		},
		{
			"adjacent - no overlap",
			TimeRange{TimeOfDay{8, 0}, TimeOfDay{10, 0}},
			[]TimeRange{{TimeOfDay{10, 0}, TimeOfDay{12, 0}}},
			false,
		},
		{
			"empty list",
			TimeRange{TimeOfDay{9, 0}, TimeOfDay{10, 0}},
			[]TimeRange{},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.Overlaps(tt.others); got != tt.want {
				t.Errorf("Overlaps() = %v, want %v", got, tt.want)
			}
		})
	}
}

// DateRange tests

func TestNewDateRange(t *testing.T) {
	start := time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)

	t.Run("valid", func(t *testing.T) {
		dr, err := NewDateRange(start, end)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !dr.Start.Equal(start) || !dr.End.Equal(end) {
			t.Errorf("got %v, want start=%v end=%v", dr, start, end)
		}
	})

	t.Run("invalid - start after end", func(t *testing.T) {
		_, err := NewDateRange(end, start)
		if err == nil {
			t.Error("expected error for start after end")
		}
	})

	t.Run("invalid - same time", func(t *testing.T) {
		_, err := NewDateRange(start, start)
		if err == nil {
			t.Error("expected error for same start and end")
		}
	})
}

func TestCalendarYear(t *testing.T) {
	dr := CalendarYear(2025)

	wantStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	if !dr.Start.Equal(wantStart) {
		t.Errorf("Start = %v, want %v", dr.Start, wantStart)
	}
	if !dr.End.Equal(wantEnd) {
		t.Errorf("End = %v, want %v", dr.End, wantEnd)
	}
}

func TestRollingYearFrom(t *testing.T) {
	start := time.Date(2025, 2, 20, 10, 30, 0, 0, time.UTC)
	dr := RollingYearFrom(start)

	wantEnd := time.Date(2026, 2, 20, 10, 30, 0, 0, time.UTC)

	if !dr.Start.Equal(start) {
		t.Errorf("Start = %v, want %v", dr.Start, start)
	}
	if !dr.End.Equal(wantEnd) {
		t.Errorf("End = %v, want %v", dr.End, wantEnd)
	}
}

func TestDateRange_Validate(t *testing.T) {
	tests := []struct {
		name    string
		dr      DateRange
		wantErr bool
	}{
		{
			"valid",
			DateRange{
				Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			},
			false,
		},
		{
			"invalid - start equals end",
			DateRange{
				Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			true,
		},
		{
			"invalid - start after end",
			DateRange{
				Start: time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dr.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDateRange_IsZero(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		dr := DateRange{}
		if !dr.IsZero() {
			t.Error("expected IsZero() = true for zero value")
		}
	})

	t.Run("non-zero start", func(t *testing.T) {
		dr := DateRange{Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
		if dr.IsZero() {
			t.Error("expected IsZero() = false when Start is set")
		}
	})

	t.Run("non-zero end", func(t *testing.T) {
		dr := DateRange{End: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
		if dr.IsZero() {
			t.Error("expected IsZero() = false when End is set")
		}
	})
}

func TestDateRange_Duration(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC)
	dr := DateRange{Start: start, End: end}

	want := 7 * 24 * time.Hour
	if got := dr.Duration(); got != want {
		t.Errorf("Duration() = %v, want %v", got, want)
	}
}

func TestDateRange_Contains(t *testing.T) {
	dr := DateRange{
		Start: time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name string
		t    time.Time
		want bool
	}{
		{"at start", time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC), true},
		{"middle", time.Date(2025, 8, 15, 12, 0, 0, 0, time.UTC), true},
		{"just before end", time.Date(2026, 2, 19, 23, 59, 59, 0, time.UTC), true},
		{"at end (exclusive)", time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC), false},
		{"before start", time.Date(2025, 2, 19, 23, 59, 59, 0, time.UTC), false},
		{"after end", time.Date(2026, 2, 21, 0, 0, 0, 0, time.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dr.Contains(tt.t); got != tt.want {
				t.Errorf("Contains(%v) = %v, want %v", tt.t, got, tt.want)
			}
		})
	}
}

func TestDateRange_ContainsInclusive(t *testing.T) {
	dr := DateRange{
		Start: time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name string
		t    time.Time
		want bool
	}{
		{"at start", time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC), true},
		{"middle", time.Date(2025, 8, 15, 12, 0, 0, 0, time.UTC), true},
		{"at end (inclusive)", time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC), true},
		{"before start", time.Date(2025, 2, 19, 23, 59, 59, 0, time.UTC), false},
		{"after end", time.Date(2026, 2, 20, 0, 0, 1, 0, time.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dr.ContainsInclusive(tt.t); got != tt.want {
				t.Errorf("ContainsInclusive(%v) = %v, want %v", tt.t, got, tt.want)
			}
		})
	}
}

func TestDateRange_Overlaps(t *testing.T) {
	dr := DateRange{
		Start: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name  string
		other DateRange
		want  bool
	}{
		{
			"no overlap - before",
			DateRange{
				Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			},
			false,
		},
		{
			"no overlap - after",
			DateRange{
				Start: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			},
			false,
		},
		{
			"overlap - partial start",
			DateRange{
				Start: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			true,
		},
		{
			"overlap - partial end",
			DateRange{
				Start: time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
			},
			true,
		},
		{
			"overlap - contained",
			DateRange{
				Start: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			true,
		},
		{
			"overlap - contains",
			DateRange{
				Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dr.Overlaps(tt.other); got != tt.want {
				t.Errorf("Overlaps(%v) = %v, want %v", tt.other, got, tt.want)
			}
		})
	}
}

func TestDateRange_NextAnnualPeriod(t *testing.T) {
	dr := DateRange{
		Start: time.Date(2025, 2, 20, 10, 30, 0, 0, time.UTC),
		End:   time.Date(2026, 2, 20, 10, 30, 0, 0, time.UTC),
	}

	next := dr.NextAnnualPeriod()

	wantStart := time.Date(2026, 2, 20, 10, 30, 0, 0, time.UTC)
	wantEnd := time.Date(2027, 2, 20, 10, 30, 0, 0, time.UTC)

	if !next.Start.Equal(wantStart) {
		t.Errorf("NextAnnualPeriod().Start = %v, want %v", next.Start, wantStart)
	}
	if !next.End.Equal(wantEnd) {
		t.Errorf("NextAnnualPeriod().End = %v, want %v", next.End, wantEnd)
	}
}

func TestDateRange_RollingPeriodScenario(t *testing.T) {
	// Test the scenario from the requirements:
	// Campaign starts 02/11/25
	// Patient A first alert: 02/20/25 → period: 02/20/25 to 02/20/26
	// Patient A hits max ~12/01/25, blocked until period expires
	// Period expires 02/20/26
	// Patient A next alert: 03/30/26 → new period: 03/30/26 to 03/30/27

	// First period
	firstAlert := time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC)
	period1 := RollingYearFrom(firstAlert)

	// Verify first period
	if !period1.Start.Equal(firstAlert) {
		t.Errorf("period1.Start = %v, want %v", period1.Start, firstAlert)
	}
	expectedEnd1 := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	if !period1.End.Equal(expectedEnd1) {
		t.Errorf("period1.End = %v, want %v", period1.End, expectedEnd1)
	}

	// Check that 12/01/25 is within period1
	dec1 := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	if !period1.Contains(dec1) {
		t.Errorf("period1 should contain %v", dec1)
	}

	// Check that 03/30/26 is NOT within period1 (period expired)
	mar30 := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)
	if period1.Contains(mar30) {
		t.Errorf("period1 should NOT contain %v (period expired)", mar30)
	}

	// New period starts from next alert
	period2 := RollingYearFrom(mar30)
	expectedEnd2 := time.Date(2027, 3, 30, 0, 0, 0, 0, time.UTC)
	if !period2.Start.Equal(mar30) {
		t.Errorf("period2.Start = %v, want %v", period2.Start, mar30)
	}
	if !period2.End.Equal(expectedEnd2) {
		t.Errorf("period2.End = %v, want %v", period2.End, expectedEnd2)
	}
}

func TestDateRange_JSON(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		dr := DateRange{
			Start: time.Date(2025, 2, 20, 10, 30, 0, 0, time.UTC),
			End:   time.Date(2026, 2, 20, 10, 30, 0, 0, time.UTC),
		}

		data, err := json.Marshal(dr)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		want := `{"Start":"2025-02-20T10:30:00Z","End":"2026-02-20T10:30:00Z"}`
		if string(data) != want {
			t.Errorf("Marshal = %s, want %s", data, want)
		}
	})

	t.Run("unmarshal", func(t *testing.T) {
		data := []byte(`{"Start":"2025-02-20T10:30:00Z","End":"2026-02-20T10:30:00Z"}`)

		var dr DateRange
		if err := json.Unmarshal(data, &dr); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		wantStart := time.Date(2025, 2, 20, 10, 30, 0, 0, time.UTC)
		wantEnd := time.Date(2026, 2, 20, 10, 30, 0, 0, time.UTC)

		if !dr.Start.Equal(wantStart) {
			t.Errorf("Start = %v, want %v", dr.Start, wantStart)
		}
		if !dr.End.Equal(wantEnd) {
			t.Errorf("End = %v, want %v", dr.End, wantEnd)
		}
	})

	t.Run("round-trip", func(t *testing.T) {
		original := DateRange{
			Start: time.Date(2025, 6, 15, 14, 30, 45, 123456789, time.UTC),
			End:   time.Date(2026, 6, 15, 14, 30, 45, 123456789, time.UTC),
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var restored DateRange
		if err := json.Unmarshal(data, &restored); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		// Go's time.Time uses RFC3339Nano which preserves nanosecond precision
		if !original.Start.Equal(restored.Start) {
			t.Errorf("Start round-trip: got %v, want %v", restored.Start, original.Start)
		}
		if !original.End.Equal(restored.End) {
			t.Errorf("End round-trip: got %v, want %v", restored.End, original.End)
		}
	})

	t.Run("unmarshal with timezone", func(t *testing.T) {
		// JSON with timezone offset
		data := []byte(`{"Start":"2025-02-20T10:30:00-05:00","End":"2026-02-20T10:30:00-05:00"}`)

		var dr DateRange
		if err := json.Unmarshal(data, &dr); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		// Should parse correctly and represent the same instant as 15:30 UTC
		wantStartUTC := time.Date(2025, 2, 20, 15, 30, 0, 0, time.UTC)
		if !dr.Start.Equal(wantStartUTC) {
			t.Errorf("Start = %v (UTC: %v), want equivalent to %v", dr.Start, dr.Start.UTC(), wantStartUTC)
		}
	})
}

func TestDateRange_TimezoneMixedComparisons(t *testing.T) {
	// Load timezone locations
	nyLoc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("Failed to load America/New_York: %v", err)
	}
	laLoc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("Failed to load America/Los_Angeles: %v", err)
	}

	t.Run("Contains with mixed timezones", func(t *testing.T) {
		// Range in UTC: 2025-06-01 00:00 to 2025-06-02 00:00
		dr := DateRange{
			Start: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC),
		}

		// 2025-06-01 12:00 NY time = 2025-06-01 16:00 UTC (within range)
		nyNoon := time.Date(2025, 6, 1, 12, 0, 0, 0, nyLoc)
		if !dr.Contains(nyNoon) {
			t.Errorf("Contains(%v) = false, want true (NY noon is within UTC day)", nyNoon)
		}

		// 2025-05-31 21:00 NY time = 2025-06-01 01:00 UTC (within range)
		nyPrevNight := time.Date(2025, 5, 31, 21, 0, 0, 0, nyLoc)
		if !dr.Contains(nyPrevNight) {
			t.Errorf("Contains(%v) = false, want true (NY 9pm May 31 = UTC 1am June 1)", nyPrevNight)
		}

		// 2025-05-31 19:00 NY time = 2025-05-31 23:00 UTC (before range)
		nyTooEarly := time.Date(2025, 5, 31, 19, 0, 0, 0, nyLoc)
		if dr.Contains(nyTooEarly) {
			t.Errorf("Contains(%v) = true, want false (NY 7pm May 31 = UTC 11pm May 31)", nyTooEarly)
		}
	})

	t.Run("Range created in one timezone, checked in another", func(t *testing.T) {
		// Create range using NY times
		drNY := DateRange{
			Start: time.Date(2025, 6, 1, 9, 0, 0, 0, nyLoc),  // 9am NY = 1pm UTC
			End:   time.Date(2025, 6, 1, 17, 0, 0, 0, nyLoc), // 5pm NY = 9pm UTC
		}

		// Check with LA time: 10am LA = 1pm NY = 5pm UTC (within range)
		laTime := time.Date(2025, 6, 1, 10, 0, 0, 0, laLoc)
		if !drNY.Contains(laTime) {
			t.Errorf("NY range Contains(LA 10am) = false, want true")
		}

		// Check with UTC: 2pm UTC (within 1pm-9pm UTC range)
		utcTime := time.Date(2025, 6, 1, 14, 0, 0, 0, time.UTC)
		if !drNY.Contains(utcTime) {
			t.Errorf("NY range Contains(UTC 2pm) = false, want true")
		}

		// Check with UTC: 10am UTC (before 1pm UTC start)
		utcEarly := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
		if drNY.Contains(utcEarly) {
			t.Errorf("NY range Contains(UTC 10am) = true, want false")
		}
	})

	t.Run("Overlaps with mixed timezones", func(t *testing.T) {
		// Range 1: UTC 12:00-18:00
		dr1 := DateRange{
			Start: time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 6, 1, 18, 0, 0, 0, time.UTC),
		}

		// Range 2: NY 10:00-14:00 = UTC 14:00-18:00 (overlaps)
		dr2 := DateRange{
			Start: time.Date(2025, 6, 1, 10, 0, 0, 0, nyLoc),
			End:   time.Date(2025, 6, 1, 14, 0, 0, 0, nyLoc),
		}

		if !dr1.Overlaps(dr2) {
			t.Errorf("UTC 12-18 should overlap with NY 10-14 (UTC 14-18)")
		}

		// Range 3: NY 6:00-8:00 = UTC 10:00-12:00 (adjacent, no overlap)
		dr3 := DateRange{
			Start: time.Date(2025, 6, 1, 6, 0, 0, 0, nyLoc),
			End:   time.Date(2025, 6, 1, 8, 0, 0, 0, nyLoc),
		}

		if dr1.Overlaps(dr3) {
			t.Errorf("UTC 12-18 should NOT overlap with NY 6-8 (UTC 10-12, adjacent)")
		}
	})

	t.Run("DST transition handling", func(t *testing.T) {
		// March 9, 2025 is when DST starts in US (clocks spring forward at 2am)
		// Create a range that spans the DST transition
		drDST := RollingYearFrom(time.Date(2025, 3, 9, 1, 30, 0, 0, nyLoc))

		// The range should still work correctly
		if drDST.Start.IsZero() || drDST.End.IsZero() {
			t.Error("DST transition range should not have zero times")
		}

		// Check a time well within the range
		midYear := time.Date(2025, 7, 1, 12, 0, 0, 0, nyLoc)
		if !drDST.Contains(midYear) {
			t.Errorf("DST range should contain mid-year date")
		}
	})
}
