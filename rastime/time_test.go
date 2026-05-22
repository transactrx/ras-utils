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
