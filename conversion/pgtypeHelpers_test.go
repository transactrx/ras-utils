package repository

import (
	"testing"
	"time"
)

func TestConvertToPgtypeString(t *testing.T) {
	t.Run("converts valid string pointer", func(t *testing.T) {
		s := "test_value"
		result := ConvertToPgtypeString(&s)

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.String != "test_value" {
			t.Errorf("expected test_value, got %s", result.String)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeString(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})

	t.Run("converts empty string", func(t *testing.T) {
		s := ""
		result := ConvertToPgtypeString(&s)

		if !result.Valid {
			t.Error("expected Valid to be true for empty string")
		}
		if result.String != "" {
			t.Errorf("expected empty string, got %s", result.String)
		}
	})
}

func TestConvertToPgtypeInt8(t *testing.T) {
	t.Run("converts valid int64 pointer", func(t *testing.T) {
		i := int64(12345)
		result := ConvertToPgtypeInt8(&i)

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Int64 != 12345 {
			t.Errorf("expected 12345, got %d", result.Int64)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeInt8(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})

	t.Run("converts zero value", func(t *testing.T) {
		i := int64(0)
		result := ConvertToPgtypeInt8(&i)

		if !result.Valid {
			t.Error("expected Valid to be true for zero")
		}
		if result.Int64 != 0 {
			t.Errorf("expected 0, got %d", result.Int64)
		}
	})
}

func TestConvertToPgtypeInt2(t *testing.T) {
	t.Run("converts valid int32 pointer", func(t *testing.T) {
		i := int32(123)
		result := ConvertToPgtypeInt2(&i)

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Int16 != 123 {
			t.Errorf("expected 123, got %d", result.Int16)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeInt2(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

func TestConvertToPgtypeBool(t *testing.T) {
	t.Run("converts true", func(t *testing.T) {
		b := true
		result := ConvertToPgtypeBool(&b)

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if !result.Bool {
			t.Error("expected Bool to be true")
		}
	})

	t.Run("converts false", func(t *testing.T) {
		b := false
		result := ConvertToPgtypeBool(&b)

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Bool {
			t.Error("expected Bool to be false")
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeBool(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

func TestConvertToPgtypeTimestamp(t *testing.T) {
	t.Run("converts valid time pointer", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result := ConvertToPgtypeTimestamp(&tm)

		if result.Time.IsZero() {
			t.Error("expected time to be set")
		}
		if !result.Time.Equal(tm) {
			t.Errorf("expected %v, got %v", tm, result.Time)
		}
	})
}
