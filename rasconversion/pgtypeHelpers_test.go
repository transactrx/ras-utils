package rasconversion

import (
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestConvertToPgtypeInt4(t *testing.T) {
	t.Run("converts valid int32 pointer", func(t *testing.T) {
		i := int32(12345)
		result := ConvertToPgtypeInt4(&i)

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Int32 != 12345 {
			t.Errorf("expected 12345, got %d", result.Int32)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeInt4(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})

	t.Run("converts negative value", func(t *testing.T) {
		i := int32(-999)
		result := ConvertToPgtypeInt4(&i)

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Int32 != -999 {
			t.Errorf("expected -999, got %d", result.Int32)
		}
	})

	t.Run("converts zero value", func(t *testing.T) {
		i := int32(0)
		result := ConvertToPgtypeInt4(&i)

		if !result.Valid {
			t.Error("expected Valid to be true for zero")
		}
		if result.Int32 != 0 {
			t.Errorf("expected 0, got %d", result.Int32)
		}
	})
}

func TestConvertToPgtypeFloat8(t *testing.T) {
	t.Run("converts valid float64 pointer", func(t *testing.T) {
		f := 123.456
		result := ConvertToPgtypeFloat8(&f)

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Float64 != 123.456 {
			t.Errorf("expected 123.456, got %f", result.Float64)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeFloat8(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})

	t.Run("converts negative value", func(t *testing.T) {
		f := -99.99
		result := ConvertToPgtypeFloat8(&f)

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Float64 != -99.99 {
			t.Errorf("expected -99.99, got %f", result.Float64)
		}
	})

	t.Run("converts zero value", func(t *testing.T) {
		f := 0.0
		result := ConvertToPgtypeFloat8(&f)

		if !result.Valid {
			t.Error("expected Valid to be true for zero")
		}
		if result.Float64 != 0.0 {
			t.Errorf("expected 0.0, got %f", result.Float64)
		}
	})
}

func TestConvertToPgtypeNumeric(t *testing.T) {
	t.Run("converts valid float64 pointer", func(t *testing.T) {
		f := 123.45
		result := ConvertToPgtypeNumeric(&f)

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeNumeric(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})

	t.Run("converts zero value", func(t *testing.T) {
		f := 0.0
		result := ConvertToPgtypeNumeric(&f)

		if !result.Valid {
			t.Error("expected Valid to be true for zero")
		}
	})
}

func TestConvertToPgtypeUUID(t *testing.T) {
	t.Run("converts valid uuid pointer", func(t *testing.T) {
		u := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		result := ConvertToPgtypeUUID(&u)

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Bytes != u {
			t.Errorf("expected %v, got %v", u, result.Bytes)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeUUID(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

func TestConvertToPgtypeUUIDFromString(t *testing.T) {
	t.Run("converts valid uuid string", func(t *testing.T) {
		s := "550e8400-e29b-41d4-a716-446655440000"
		result := ConvertToPgtypeUUIDFromString(&s)

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		expected := uuid.MustParse(s)
		if result.Bytes != expected {
			t.Errorf("expected %v, got %v", expected, result.Bytes)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeUUIDFromString(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})

	t.Run("handles invalid uuid string", func(t *testing.T) {
		s := "not-a-uuid"
		result := ConvertToPgtypeUUIDFromString(&s)

		if result.Valid {
			t.Error("expected Valid to be false for invalid uuid string")
		}
	})
}

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

func TestTryConvertToPgtypeInt4(t *testing.T) {
	t.Run("converts valid int32 pointer", func(t *testing.T) {
		i := int32(12345)
		result, err := TryConvertToPgtypeInt4(&i)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Int32 != 12345 {
			t.Errorf("expected 12345, got %d", result.Int32)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeInt4(nil)

		if err != nil {
			t.Errorf("unexpected error for nil: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

func TestTryConvertToPgtypeFloat8(t *testing.T) {
	t.Run("converts valid float64 pointer", func(t *testing.T) {
		f := 123.456
		result, err := TryConvertToPgtypeFloat8(&f)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Float64 != 123.456 {
			t.Errorf("expected 123.456, got %f", result.Float64)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeFloat8(nil)

		if err != nil {
			t.Errorf("unexpected error for nil: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

func TestTryConvertToPgtypeNumeric(t *testing.T) {
	t.Run("converts valid float64 pointer", func(t *testing.T) {
		f := 123.45
		result, err := TryConvertToPgtypeNumeric(&f)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeNumeric(nil)

		if err != nil {
			t.Errorf("unexpected error for nil: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

func TestTryConvertToPgtypeUUID(t *testing.T) {
	t.Run("converts valid uuid pointer", func(t *testing.T) {
		u := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		result, err := TryConvertToPgtypeUUID(&u)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Bytes != u {
			t.Errorf("expected %v, got %v", u, result.Bytes)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeUUID(nil)

		if err != nil {
			t.Errorf("unexpected error for nil: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

func TestTryConvertToPgtypeUUIDFromString(t *testing.T) {
	t.Run("converts valid uuid string", func(t *testing.T) {
		s := "550e8400-e29b-41d4-a716-446655440000"
		result, err := TryConvertToPgtypeUUIDFromString(&s)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeUUIDFromString(nil)

		if err != nil {
			t.Errorf("unexpected error for nil: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})

	t.Run("returns error for invalid uuid string", func(t *testing.T) {
		s := "not-a-uuid"
		_, err := TryConvertToPgtypeUUIDFromString(&s)

		if err == nil {
			t.Error("expected error for invalid uuid string")
		}
	})
}

// Tests for reverse conversions (pgtype to Go pointer)

func TestConvertFromPgtypeText(t *testing.T) {
	t.Run("converts valid text", func(t *testing.T) {
		pt := pgtype.Text{String: "hello", Valid: true}
		result := ConvertFromPgtypeText(pt)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if *result != "hello" {
			t.Errorf("expected hello, got %s", *result)
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pt := pgtype.Text{Valid: false}
		result := ConvertFromPgtypeText(pt)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})

	t.Run("converts empty string", func(t *testing.T) {
		pt := pgtype.Text{String: "", Valid: true}
		result := ConvertFromPgtypeText(pt)

		if result == nil {
			t.Error("expected non-nil result for empty string")
		}
		if *result != "" {
			t.Errorf("expected empty string, got %s", *result)
		}
	})
}

func TestConvertFromPgtypeInt2(t *testing.T) {
	t.Run("converts valid int2", func(t *testing.T) {
		pi := pgtype.Int2{Int16: 123, Valid: true}
		result := ConvertFromPgtypeInt2(pi)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if *result != 123 {
			t.Errorf("expected 123, got %d", *result)
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pi := pgtype.Int2{Valid: false}
		result := ConvertFromPgtypeInt2(pi)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})
}

func TestConvertFromPgtypeInt4(t *testing.T) {
	t.Run("converts valid int4", func(t *testing.T) {
		pi := pgtype.Int4{Int32: 12345, Valid: true}
		result := ConvertFromPgtypeInt4(pi)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if *result != 12345 {
			t.Errorf("expected 12345, got %d", *result)
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pi := pgtype.Int4{Valid: false}
		result := ConvertFromPgtypeInt4(pi)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})

	t.Run("converts negative value", func(t *testing.T) {
		pi := pgtype.Int4{Int32: -999, Valid: true}
		result := ConvertFromPgtypeInt4(pi)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if *result != -999 {
			t.Errorf("expected -999, got %d", *result)
		}
	})
}

func TestConvertFromPgtypeInt8(t *testing.T) {
	t.Run("converts valid int8", func(t *testing.T) {
		pi := pgtype.Int8{Int64: 123456789, Valid: true}
		result := ConvertFromPgtypeInt8(pi)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if *result != 123456789 {
			t.Errorf("expected 123456789, got %d", *result)
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pi := pgtype.Int8{Valid: false}
		result := ConvertFromPgtypeInt8(pi)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})
}

func TestConvertFromPgtypeBool(t *testing.T) {
	t.Run("converts true", func(t *testing.T) {
		pb := pgtype.Bool{Bool: true, Valid: true}
		result := ConvertFromPgtypeBool(pb)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if *result != true {
			t.Error("expected true")
		}
	})

	t.Run("converts false", func(t *testing.T) {
		pb := pgtype.Bool{Bool: false, Valid: true}
		result := ConvertFromPgtypeBool(pb)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if *result != false {
			t.Error("expected false")
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pb := pgtype.Bool{Valid: false}
		result := ConvertFromPgtypeBool(pb)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})
}

func TestConvertFromPgtypeFloat8(t *testing.T) {
	t.Run("converts valid float8", func(t *testing.T) {
		pf := pgtype.Float8{Float64: 123.456, Valid: true}
		result := ConvertFromPgtypeFloat8(pf)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if *result != 123.456 {
			t.Errorf("expected 123.456, got %f", *result)
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pf := pgtype.Float8{Valid: false}
		result := ConvertFromPgtypeFloat8(pf)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})
}

func TestConvertFromPgtypeNumeric(t *testing.T) {
	t.Run("converts valid numeric", func(t *testing.T) {
		var pn pgtype.Numeric
		_ = pn.Scan("123.45")
		result := ConvertFromPgtypeNumeric(pn)

		if result == nil {
			t.Error("expected non-nil result")
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pn := pgtype.Numeric{Valid: false}
		result := ConvertFromPgtypeNumeric(pn)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})
}

func TestConvertFromPgtypeNumericToBigFloat(t *testing.T) {
	t.Run("converts valid numeric", func(t *testing.T) {
		var pn pgtype.Numeric
		_ = pn.Scan("123.45")
		result := ConvertFromPgtypeNumericToBigFloat(pn)

		if result == nil {
			t.Error("expected non-nil result")
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pn := pgtype.Numeric{Valid: false}
		result := ConvertFromPgtypeNumericToBigFloat(pn)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})
}

func TestConvertFromPgtypeTimestamp(t *testing.T) {
	t.Run("converts valid timestamp", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		pt := pgtype.Timestamp{Time: tm, Valid: true}
		result := ConvertFromPgtypeTimestamp(pt)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if !result.Equal(tm) {
			t.Errorf("expected %v, got %v", tm, *result)
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pt := pgtype.Timestamp{Valid: false}
		result := ConvertFromPgtypeTimestamp(pt)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})
}

func TestConvertFromPgtypeTimestamptz(t *testing.T) {
	t.Run("converts valid timestamptz", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		pt := pgtype.Timestamptz{Time: tm, Valid: true}
		result := ConvertFromPgtypeTimestamptz(pt)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if !result.Equal(tm) {
			t.Errorf("expected %v, got %v", tm, *result)
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pt := pgtype.Timestamptz{Valid: false}
		result := ConvertFromPgtypeTimestamptz(pt)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})
}

func TestConvertFromPgtypeDate(t *testing.T) {
	t.Run("converts valid date", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		pd := pgtype.Date{Time: tm, Valid: true}
		result := ConvertFromPgtypeDate(pd)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if !result.Equal(tm) {
			t.Errorf("expected %v, got %v", tm, *result)
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pd := pgtype.Date{Valid: false}
		result := ConvertFromPgtypeDate(pd)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})
}

func TestConvertFromPgtypeTime(t *testing.T) {
	t.Run("converts valid time", func(t *testing.T) {
		pt := pgtype.Time{Microseconds: 10*3600_000_000 + 30*60_000_000 + 45*1_000_000, Valid: true}
		result := ConvertFromPgtypeTime(pt)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if result.Hour() != 10 {
			t.Errorf("expected hour 10, got %d", result.Hour())
		}
		if result.Minute() != 30 {
			t.Errorf("expected minute 30, got %d", result.Minute())
		}
		if result.Second() != 45 {
			t.Errorf("expected second 45, got %d", result.Second())
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pt := pgtype.Time{Valid: false}
		result := ConvertFromPgtypeTime(pt)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})
}

func TestConvertFromPgtypeUUID(t *testing.T) {
	t.Run("converts valid uuid", func(t *testing.T) {
		u := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		pu := pgtype.UUID{Bytes: u, Valid: true}
		result := ConvertFromPgtypeUUID(pu)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if *result != u {
			t.Errorf("expected %v, got %v", u, *result)
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pu := pgtype.UUID{Valid: false}
		result := ConvertFromPgtypeUUID(pu)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})
}

func TestConvertFromPgtypeUUIDToString(t *testing.T) {
	t.Run("converts valid uuid to string", func(t *testing.T) {
		u := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		pu := pgtype.UUID{Bytes: u, Valid: true}
		result := ConvertFromPgtypeUUIDToString(pu)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if *result != "550e8400-e29b-41d4-a716-446655440000" {
			t.Errorf("expected 550e8400-e29b-41d4-a716-446655440000, got %s", *result)
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pu := pgtype.UUID{Valid: false}
		result := ConvertFromPgtypeUUIDToString(pu)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})
}

// Tests for Float4, Interval, JSONB

func TestConvertToPgtypeFloat4(t *testing.T) {
	t.Run("converts valid float32 pointer", func(t *testing.T) {
		f := float32(123.456)
		result := ConvertToPgtypeFloat4(&f)

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Float32 != 123.456 {
			t.Errorf("expected 123.456, got %f", result.Float32)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeFloat4(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

func TestConvertToPgtypeInterval(t *testing.T) {
	t.Run("converts valid duration pointer", func(t *testing.T) {
		d := 2*time.Hour + 30*time.Minute
		result := ConvertToPgtypeInterval(&d)

		if !result.Valid {
			t.Error("expected Valid to be true")
		}
		if result.Microseconds != d.Microseconds() {
			t.Errorf("expected %d microseconds, got %d", d.Microseconds(), result.Microseconds)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result := ConvertToPgtypeInterval(nil)

		if result.Valid {
			t.Error("expected Valid to be false for nil pointer")
		}
	})
}

func TestConvertToPgtypeJSONB(t *testing.T) {
	t.Run("converts struct to JSONB", func(t *testing.T) {
		type testStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		v := testStruct{Name: "test", Age: 25}
		result := ConvertToPgtypeJSONB(v)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if string(result) != `{"name":"test","age":25}` {
			t.Errorf("unexpected JSON: %s", string(result))
		}
	})

	t.Run("converts map to JSONB", func(t *testing.T) {
		v := map[string]int{"a": 1, "b": 2}
		result := ConvertToPgtypeJSONB(v)

		if result == nil {
			t.Error("expected non-nil result")
		}
	})

	t.Run("handles nil", func(t *testing.T) {
		result := ConvertToPgtypeJSONB(nil)

		if result != nil {
			t.Error("expected nil for nil input")
		}
	})
}

func TestTryConvertToPgtypeJSONB(t *testing.T) {
	t.Run("converts struct to JSONB", func(t *testing.T) {
		type testStruct struct {
			Name string `json:"name"`
		}
		v := testStruct{Name: "test"}
		result, err := TryConvertToPgtypeJSONB(v)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected non-nil result")
		}
	})

	t.Run("handles nil", func(t *testing.T) {
		result, err := TryConvertToPgtypeJSONB(nil)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != nil {
			t.Error("expected nil for nil input")
		}
	})
}

func TestConvertFromPgtypeFloat4(t *testing.T) {
	t.Run("converts valid float4", func(t *testing.T) {
		pf := pgtype.Float4{Float32: 123.456, Valid: true}
		result := ConvertFromPgtypeFloat4(pf)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if *result != 123.456 {
			t.Errorf("expected 123.456, got %f", *result)
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pf := pgtype.Float4{Valid: false}
		result := ConvertFromPgtypeFloat4(pf)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})
}

func TestConvertFromPgtypeInterval(t *testing.T) {
	t.Run("converts valid interval", func(t *testing.T) {
		pi := pgtype.Interval{Microseconds: 9000000000, Valid: true} // 2.5 hours
		result := ConvertFromPgtypeInterval(pi)

		if result == nil {
			t.Error("expected non-nil result")
		}
		expected := 2*time.Hour + 30*time.Minute
		if *result != expected {
			t.Errorf("expected %v, got %v", expected, *result)
		}
	})

	t.Run("handles days", func(t *testing.T) {
		pi := pgtype.Interval{Days: 2, Valid: true}
		result := ConvertFromPgtypeInterval(pi)

		if result == nil {
			t.Error("expected non-nil result")
		}
		expected := 48 * time.Hour
		if *result != expected {
			t.Errorf("expected %v, got %v", expected, *result)
		}
	})

	t.Run("returns nil for invalid", func(t *testing.T) {
		pi := pgtype.Interval{Valid: false}
		result := ConvertFromPgtypeInterval(pi)

		if result != nil {
			t.Error("expected nil for invalid")
		}
	})
}

func TestConvertFromPgtypeJSONB(t *testing.T) {
	t.Run("unmarshals to struct", func(t *testing.T) {
		type testStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		b := []byte(`{"name":"test","age":25}`)
		result := ConvertFromPgtypeJSONB[testStruct](b)

		if result == nil {
			t.Error("expected non-nil result")
		}
		if result.Name != "test" {
			t.Errorf("expected name 'test', got '%s'", result.Name)
		}
		if result.Age != 25 {
			t.Errorf("expected age 25, got %d", result.Age)
		}
	})

	t.Run("handles nil", func(t *testing.T) {
		type testStruct struct{}
		result := ConvertFromPgtypeJSONB[testStruct](nil)

		if result != nil {
			t.Error("expected nil for nil input")
		}
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		type testStruct struct{}
		result := ConvertFromPgtypeJSONB[testStruct]([]byte("invalid"))

		if result != nil {
			t.Error("expected nil for invalid JSON")
		}
	})
}

func TestTryConvertFromPgtypeJSONB(t *testing.T) {
	t.Run("unmarshals to struct", func(t *testing.T) {
		type testStruct struct {
			Name string `json:"name"`
		}
		b := []byte(`{"name":"test"}`)
		result, err := TryConvertFromPgtypeJSONB[testStruct](b)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected non-nil result")
		}
		if result.Name != "test" {
			t.Errorf("expected name 'test', got '%s'", result.Name)
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		type testStruct struct{}
		_, err := TryConvertFromPgtypeJSONB[testStruct]([]byte("invalid"))

		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

func TestTryConvertToPgtypeFloat4(t *testing.T) {
	t.Run("converts valid float32 pointer", func(t *testing.T) {
		f := float32(123.456)
		result, err := TryConvertToPgtypeFloat4(&f)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeFloat4(nil)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false")
		}
	})
}

func TestTryConvertToPgtypeInterval(t *testing.T) {
	t.Run("converts valid duration pointer", func(t *testing.T) {
		d := time.Hour
		result, err := TryConvertToPgtypeInterval(&d)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.Valid {
			t.Error("expected Valid to be true")
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		result, err := TryConvertToPgtypeInterval(nil)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.Valid {
			t.Error("expected Valid to be false")
		}
	})
}

// Tests for OrDefault variants

func TestConvertFromPgtypeTextOrDefault(t *testing.T) {
	t.Run("returns value when valid", func(t *testing.T) {
		pt := pgtype.Text{String: "hello", Valid: true}
		result := ConvertFromPgtypeTextOrDefault(pt, "default")

		if result != "hello" {
			t.Errorf("expected 'hello', got '%s'", result)
		}
	})

	t.Run("returns default when invalid", func(t *testing.T) {
		pt := pgtype.Text{Valid: false}
		result := ConvertFromPgtypeTextOrDefault(pt, "default")

		if result != "default" {
			t.Errorf("expected 'default', got '%s'", result)
		}
	})
}

func TestConvertFromPgtypeInt2OrDefault(t *testing.T) {
	t.Run("returns value when valid", func(t *testing.T) {
		pi := pgtype.Int2{Int16: 123, Valid: true}
		result := ConvertFromPgtypeInt2OrDefault(pi, 0)

		if result != 123 {
			t.Errorf("expected 123, got %d", result)
		}
	})

	t.Run("returns default when invalid", func(t *testing.T) {
		pi := pgtype.Int2{Valid: false}
		result := ConvertFromPgtypeInt2OrDefault(pi, -1)

		if result != -1 {
			t.Errorf("expected -1, got %d", result)
		}
	})
}

func TestConvertFromPgtypeInt4OrDefault(t *testing.T) {
	t.Run("returns value when valid", func(t *testing.T) {
		pi := pgtype.Int4{Int32: 12345, Valid: true}
		result := ConvertFromPgtypeInt4OrDefault(pi, 0)

		if result != 12345 {
			t.Errorf("expected 12345, got %d", result)
		}
	})

	t.Run("returns default when invalid", func(t *testing.T) {
		pi := pgtype.Int4{Valid: false}
		result := ConvertFromPgtypeInt4OrDefault(pi, -1)

		if result != -1 {
			t.Errorf("expected -1, got %d", result)
		}
	})
}

func TestConvertFromPgtypeInt8OrDefault(t *testing.T) {
	t.Run("returns value when valid", func(t *testing.T) {
		pi := pgtype.Int8{Int64: 123456789, Valid: true}
		result := ConvertFromPgtypeInt8OrDefault(pi, 0)

		if result != 123456789 {
			t.Errorf("expected 123456789, got %d", result)
		}
	})

	t.Run("returns default when invalid", func(t *testing.T) {
		pi := pgtype.Int8{Valid: false}
		result := ConvertFromPgtypeInt8OrDefault(pi, -1)

		if result != -1 {
			t.Errorf("expected -1, got %d", result)
		}
	})
}

func TestConvertFromPgtypeBoolOrDefault(t *testing.T) {
	t.Run("returns value when valid", func(t *testing.T) {
		pb := pgtype.Bool{Bool: true, Valid: true}
		result := ConvertFromPgtypeBoolOrDefault(pb, false)

		if result != true {
			t.Error("expected true")
		}
	})

	t.Run("returns default when invalid", func(t *testing.T) {
		pb := pgtype.Bool{Valid: false}
		result := ConvertFromPgtypeBoolOrDefault(pb, true)

		if result != true {
			t.Error("expected default true")
		}
	})
}

func TestConvertFromPgtypeFloat4OrDefault(t *testing.T) {
	t.Run("returns value when valid", func(t *testing.T) {
		pf := pgtype.Float4{Float32: 123.456, Valid: true}
		result := ConvertFromPgtypeFloat4OrDefault(pf, 0)

		if result != 123.456 {
			t.Errorf("expected 123.456, got %f", result)
		}
	})

	t.Run("returns default when invalid", func(t *testing.T) {
		pf := pgtype.Float4{Valid: false}
		result := ConvertFromPgtypeFloat4OrDefault(pf, -1.0)

		if result != -1.0 {
			t.Errorf("expected -1.0, got %f", result)
		}
	})
}

func TestConvertFromPgtypeFloat8OrDefault(t *testing.T) {
	t.Run("returns value when valid", func(t *testing.T) {
		pf := pgtype.Float8{Float64: 123.456, Valid: true}
		result := ConvertFromPgtypeFloat8OrDefault(pf, 0)

		if result != 123.456 {
			t.Errorf("expected 123.456, got %f", result)
		}
	})

	t.Run("returns default when invalid", func(t *testing.T) {
		pf := pgtype.Float8{Valid: false}
		result := ConvertFromPgtypeFloat8OrDefault(pf, -1.0)

		if result != -1.0 {
			t.Errorf("expected -1.0, got %f", result)
		}
	})
}

func TestConvertFromPgtypeNumericOrDefault(t *testing.T) {
	t.Run("returns value when valid", func(t *testing.T) {
		var pn pgtype.Numeric
		_ = pn.Scan("123.45")
		result := ConvertFromPgtypeNumericOrDefault(pn, 0)

		if result < 123.44 || result > 123.46 {
			t.Errorf("expected ~123.45, got %f", result)
		}
	})

	t.Run("returns default when invalid", func(t *testing.T) {
		pn := pgtype.Numeric{Valid: false}
		result := ConvertFromPgtypeNumericOrDefault(pn, -1.0)

		if result != -1.0 {
			t.Errorf("expected -1.0, got %f", result)
		}
	})
}

func TestConvertFromPgtypeTimestampOrDefault(t *testing.T) {
	t.Run("returns value when valid", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		pt := pgtype.Timestamp{Time: tm, Valid: true}
		def := time.Time{}
		result := ConvertFromPgtypeTimestampOrDefault(pt, def)

		if !result.Equal(tm) {
			t.Errorf("expected %v, got %v", tm, result)
		}
	})

	t.Run("returns default when invalid", func(t *testing.T) {
		pt := pgtype.Timestamp{Valid: false}
		def := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		result := ConvertFromPgtypeTimestampOrDefault(pt, def)

		if !result.Equal(def) {
			t.Errorf("expected %v, got %v", def, result)
		}
	})
}

func TestConvertFromPgtypeTimestamptzOrDefault(t *testing.T) {
	t.Run("returns value when valid", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		pt := pgtype.Timestamptz{Time: tm, Valid: true}
		def := time.Time{}
		result := ConvertFromPgtypeTimestamptzOrDefault(pt, def)

		if !result.Equal(tm) {
			t.Errorf("expected %v, got %v", tm, result)
		}
	})

	t.Run("returns default when invalid", func(t *testing.T) {
		pt := pgtype.Timestamptz{Valid: false}
		def := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		result := ConvertFromPgtypeTimestamptzOrDefault(pt, def)

		if !result.Equal(def) {
			t.Errorf("expected %v, got %v", def, result)
		}
	})
}

func TestConvertFromPgtypeDateOrDefault(t *testing.T) {
	t.Run("returns value when valid", func(t *testing.T) {
		tm := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		pd := pgtype.Date{Time: tm, Valid: true}
		def := time.Time{}
		result := ConvertFromPgtypeDateOrDefault(pd, def)

		if !result.Equal(tm) {
			t.Errorf("expected %v, got %v", tm, result)
		}
	})

	t.Run("returns default when invalid", func(t *testing.T) {
		pd := pgtype.Date{Valid: false}
		def := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		result := ConvertFromPgtypeDateOrDefault(pd, def)

		if !result.Equal(def) {
			t.Errorf("expected %v, got %v", def, result)
		}
	})
}

func TestConvertFromPgtypeIntervalOrDefault(t *testing.T) {
	t.Run("returns value when valid", func(t *testing.T) {
		pi := pgtype.Interval{Microseconds: 3600000000, Valid: true} // 1 hour
		def := time.Duration(0)
		result := ConvertFromPgtypeIntervalOrDefault(pi, def)

		if result != time.Hour {
			t.Errorf("expected %v, got %v", time.Hour, result)
		}
	})

	t.Run("returns default when invalid", func(t *testing.T) {
		pi := pgtype.Interval{Valid: false}
		def := time.Minute
		result := ConvertFromPgtypeIntervalOrDefault(pi, def)

		if result != def {
			t.Errorf("expected %v, got %v", def, result)
		}
	})
}

func TestConvertFromPgtypeUUIDOrDefault(t *testing.T) {
	t.Run("returns value when valid", func(t *testing.T) {
		u := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		pu := pgtype.UUID{Bytes: u, Valid: true}
		def := uuid.Nil
		result := ConvertFromPgtypeUUIDOrDefault(pu, def)

		if result != u {
			t.Errorf("expected %v, got %v", u, result)
		}
	})

	t.Run("returns default when invalid", func(t *testing.T) {
		pu := pgtype.UUID{Valid: false}
		def := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		result := ConvertFromPgtypeUUIDOrDefault(pu, def)

		if result != def {
			t.Errorf("expected %v, got %v", def, result)
		}
	})
}

func TestConvertFromPgtypeUUIDToStringOrDefault(t *testing.T) {
	t.Run("returns value when valid", func(t *testing.T) {
		u := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		pu := pgtype.UUID{Bytes: u, Valid: true}
		result := ConvertFromPgtypeUUIDToStringOrDefault(pu, "default")

		if result != "550e8400-e29b-41d4-a716-446655440000" {
			t.Errorf("expected UUID string, got '%s'", result)
		}
	})

	t.Run("returns default when invalid", func(t *testing.T) {
		pu := pgtype.UUID{Valid: false}
		result := ConvertFromPgtypeUUIDToStringOrDefault(pu, "default")

		if result != "default" {
			t.Errorf("expected 'default', got '%s'", result)
		}
	})
}

// Suppress unused import warnings
var _ = big.Float{}
