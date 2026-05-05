package rasconversion

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

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if !result.Time.Equal(tm) {
			t.Errorf("expected %v, got %v", tm, result.Time)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeTimestamp(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

func TestConvertToPgtypeTimestamptz(t *testing.T) {
	t.Run("converts valid time pointer", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result := ConvertToPgtypeTimestamptz(&tm)

		if !result.Valid {
			t.Error("expected Valid to be true for valid time pointer")
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeTimestamptz(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

func TestConvertToPgtypeDate(t *testing.T) {
	t.Run("converts valid pointer", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result := ConvertToPgtypeDate(&tm)

		if !result.Valid {
			t.Error("expected Valid to be true for valid pointer")
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeDate(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

func TestConvertToPgtypeTime(t *testing.T) {
	t.Run("converts valid pointer", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result := ConvertToPgtypeTime(&tm)

		if !result.Valid {
			t.Error("expected Valid to be true for valid pointer")
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeTime(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

// Tests for TryConvert* error-returning variants

func TestTryConvertToPgtypeString(t *testing.T) {
	t.Run("converts valid string pointer", func(t *testing.T) {
		s := "test_value"
		result, err := TryConvertToPgtypeString(&s)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.String != "test_value" {
			t.Errorf("expected test_value, got %s", result.String)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeString(nil)

		if err != nil {
			t.Errorf("unexpected error for nil: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})

	t.Run("converts empty string", func(t *testing.T) {
		s := ""
		result, err := TryConvertToPgtypeString(&s)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true for empty string")
		}
	})
}

func TestTryConvertToPgtypeInt8(t *testing.T) {
	t.Run("converts valid int64 pointer", func(t *testing.T) {
		i := int64(12345)
		result, err := TryConvertToPgtypeInt8(&i)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Int64 != 12345 {
			t.Errorf("expected 12345, got %d", result.Int64)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeInt8(nil)

		if err != nil {
			t.Errorf("unexpected error for nil: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})

	t.Run("converts negative value", func(t *testing.T) {
		i := int64(-999)
		result, err := TryConvertToPgtypeInt8(&i)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Int64 != -999 {
			t.Errorf("expected -999, got %d", result.Int64)
		}
	})
}

func TestTryConvertToPgtypeInt2(t *testing.T) {
	t.Run("converts valid int32 pointer", func(t *testing.T) {
		i := int32(123)
		result, err := TryConvertToPgtypeInt2(&i)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Int16 != 123 {
			t.Errorf("expected 123, got %d", result.Int16)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeInt2(nil)

		if err != nil {
			t.Errorf("unexpected error for nil: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})

	t.Run("converts negative value", func(t *testing.T) {
		i := int32(-50)
		result, err := TryConvertToPgtypeInt2(&i)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Int16 != -50 {
			t.Errorf("expected -50, got %d", result.Int16)
		}
	})
}

func TestTryConvertToPgtypeBool(t *testing.T) {
	t.Run("converts true", func(t *testing.T) {
		b := true
		result, err := TryConvertToPgtypeBool(&b)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if !result.Bool {
			t.Error("expected Bool to be true")
		}
	})

	t.Run("converts false", func(t *testing.T) {
		b := false
		result, err := TryConvertToPgtypeBool(&b)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Bool {
			t.Error("expected Bool to be false")
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeBool(nil)

		if err != nil {
			t.Errorf("unexpected error for nil: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

func TestTryConvertToPgtypeTimestamp(t *testing.T) {
	t.Run("converts valid time pointer", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result, err := TryConvertToPgtypeTimestamp(&tm)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if !result.Time.Equal(tm) {
			t.Errorf("expected %v, got %v", tm, result.Time)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeTimestamp(nil)

		if err != nil {
			t.Errorf("unexpected error for nil: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})

	t.Run("converts zero time", func(t *testing.T) {
		tm := time.Time{}
		result, err := TryConvertToPgtypeTimestamp(&tm)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true for zero time")
		}
	})
}

func TestTryConvertToPgtypeTimestamptz(t *testing.T) {
	t.Run("converts valid time pointer", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result, err := TryConvertToPgtypeTimestamptz(&tm)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if !result.Time.Equal(tm) {
			t.Errorf("expected %v, got %v", tm, result.Time)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeTimestamptz(nil)

		if err != nil {
			t.Errorf("unexpected error for nil: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})

	t.Run("converts zero time", func(t *testing.T) {
		tm := time.Time{}
		result, err := TryConvertToPgtypeTimestamptz(&tm)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true for zero time")
		}
	})
}

func TestTryConvertToPgtypeDate(t *testing.T) {
	t.Run("converts valid pointer", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result, err := TryConvertToPgtypeDate(&tm)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true for valid pointer")
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeDate(nil)

		if err != nil {
			t.Errorf("unexpected error for nil: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

func TestTryConvertToPgtypeTime(t *testing.T) {
	t.Run("converts valid pointer", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result, err := TryConvertToPgtypeTime(&tm)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true for valid pointer")
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeTime(nil)

		if err != nil {
			t.Errorf("unexpected error for nil: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}
