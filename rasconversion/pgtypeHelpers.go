package rasconversion

import (
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/transactrx/ras-utils/rastime"
)

// ConvertToPgtypeString converts a nullable string to pgtype.Text. Returns invalid if nil or on error.
func ConvertToPgtypeString(s *string) pgtype.Text {
	var pgString pgtype.Text
	if s != nil {
		err := pgString.Scan(*s)
		if err != nil {
			slog.Error("failed to convert to pgtype.Text", "error", err, "string", s)
			pgString = pgtype.Text{Valid: false}
		}
	} else {
		pgString = pgtype.Text{Valid: false}
	}
	return pgString
}

// ConvertToPgtypeInt8 converts a nullable int64 to pgtype.Int8. Returns invalid if nil or on error.
func ConvertToPgtypeInt8(i *int64) pgtype.Int8 {
	var pgInt8 pgtype.Int8
	if i != nil {
		err := pgInt8.Scan(*i)
		if err != nil {
			slog.Error("failed to convert to pgtype.Int8", "error", err, "int", i)
			pgInt8 = pgtype.Int8{Valid: false}
		}
	} else {
		pgInt8 = pgtype.Int8{Valid: false}
	}
	return pgInt8
}

// ConvertToPgtypeInt2 converts a nullable int32 to pgtype.Int2. Returns invalid if nil.
func ConvertToPgtypeInt2(i *int32) pgtype.Int2 {
	if i != nil {
		// Directly construct pgtype.Int2 with int16 value
		return pgtype.Int2{
			Int16: int16(*i),
			Valid: true,
		}
	}
	return pgtype.Int2{Valid: false}
}

// ConvertToPgtypeBool converts a nullable bool to pgtype.Bool. Returns invalid if nil or on error.
func ConvertToPgtypeBool(b *bool) pgtype.Bool {
	var pgBool pgtype.Bool
	if b != nil {
		err := pgBool.Scan(*b)
		if err != nil {
			slog.Error("failed to convert to pgtype.Bool", "error", err, "bool", b)
			pgBool = pgtype.Bool{Valid: false}
		}
	} else {
		pgBool = pgtype.Bool{Valid: false}
	}
	return pgBool
}

// ConvertToPgtypeTimestamp converts a nullable time.Time to pgtype.Timestamp. Returns invalid if nil or on error.
func ConvertToPgtypeTimestamp(t *time.Time) pgtype.Timestamp {
	var pgTime pgtype.Timestamp
	if t != nil {
		err := pgTime.Scan(*t)
		if err != nil {
			slog.Error("failed to convert to pgtype.Timestamp", "error", err, "time", t)
			pgTime = pgtype.Timestamp{Valid: false}
		}
	} else {
		pgTime = pgtype.Timestamp{Valid: false}
	}
	return pgTime
}

// ConvertToPgtypeTimestamptz converts a nullable time.Time to pgtype.Timestamptz. Returns invalid if nil or on error.
func ConvertToPgtypeTimestamptz(t *time.Time) pgtype.Timestamptz {
	var pgTimez pgtype.Timestamptz
	if t != nil {
		err := pgTimez.Scan(*t)
		if err != nil {
			slog.Error("failed to convert to pgtype.Timestamptz", "error", err, "time", t)
			pgTimez = pgtype.Timestamptz{Valid: false}
		}
	} else {
		pgTimez = pgtype.Timestamptz{Valid: false}
	}
	return pgTimez
}

// ConvertToPgtypeDate converts a nullable time.Time to pgtype.Date. Returns invalid if nil or on error.
func ConvertToPgtypeDate(d *time.Time) pgtype.Date {
	var pgDate pgtype.Date
	if d != nil {
		err := pgDate.Scan(*d)
		if err != nil {
			slog.Error("failed to convert to pgtype.Date", "error", err, "time", d)
			pgDate = pgtype.Date{Valid: false}
		}
	} else {
		pgDate = pgtype.Date{Valid: false}
	}
	return pgDate
}

// ConvertToPgtypeTime converts a nullable time.Time to pgtype.Time (time of day). Returns invalid if nil.
func ConvertToPgtypeTime(t *time.Time) pgtype.Time {
	if t == nil {
		return pgtype.Time{Valid: false}
	}
	// pgtype.Time stores microseconds since midnight
	usec := int64(t.Hour())*3600_000_000 +
		int64(t.Minute())*60_000_000 +
		int64(t.Second())*1_000_000 +
		int64(t.Nanosecond())/1000
	return pgtype.Time{Microseconds: usec, Valid: true}
}

// Error-returning variants for callers that need explicit error handling

// TryConvertToPgtypeString converts a nullable string to pgtype.Text, returning an error on failure.
func TryConvertToPgtypeString(s *string) (pgtype.Text, error) {
	if s == nil {
		return pgtype.Text{Valid: false}, nil
	}
	var pgString pgtype.Text
	if err := pgString.Scan(*s); err != nil {
		return pgtype.Text{Valid: false}, err
	}
	return pgString, nil
}

// TryConvertToPgtypeInt8 converts a nullable int64 to pgtype.Int8, returning an error on failure.
func TryConvertToPgtypeInt8(i *int64) (pgtype.Int8, error) {
	if i == nil {
		return pgtype.Int8{Valid: false}, nil
	}
	var pgInt8 pgtype.Int8
	if err := pgInt8.Scan(*i); err != nil {
		return pgtype.Int8{Valid: false}, err
	}
	return pgInt8, nil
}

// TryConvertToPgtypeInt2 converts a nullable int32 to pgtype.Int2, returning an error on failure.
func TryConvertToPgtypeInt2(i *int32) (pgtype.Int2, error) {
	if i == nil {
		return pgtype.Int2{Valid: false}, nil
	}
	return pgtype.Int2{
		Int16: int16(*i),
		Valid: true,
	}, nil
}

// TryConvertToPgtypeBool converts a nullable bool to pgtype.Bool, returning an error on failure.
func TryConvertToPgtypeBool(b *bool) (pgtype.Bool, error) {
	if b == nil {
		return pgtype.Bool{Valid: false}, nil
	}
	var pgBool pgtype.Bool
	if err := pgBool.Scan(*b); err != nil {
		return pgtype.Bool{Valid: false}, err
	}
	return pgBool, nil
}

// TryConvertToPgtypeTimestamp converts a nullable time.Time to pgtype.Timestamp, returning an error on failure.
func TryConvertToPgtypeTimestamp(t *time.Time) (pgtype.Timestamp, error) {
	if t == nil {
		return pgtype.Timestamp{Valid: false}, nil
	}
	var pgTime pgtype.Timestamp
	if err := pgTime.Scan(*t); err != nil {
		return pgtype.Timestamp{Valid: false}, err
	}
	return pgTime, nil
}

// TryConvertToPgtypeTimestamptz converts a nullable time.Time to pgtype.Timestamptz, returning an error on failure.
func TryConvertToPgtypeTimestamptz(t *time.Time) (pgtype.Timestamptz, error) {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}, nil
	}
	var pgTimez pgtype.Timestamptz
	if err := pgTimez.Scan(*t); err != nil {
		return pgtype.Timestamptz{Valid: false}, err
	}
	return pgTimez, nil
}

// TryConvertToPgtypeDate converts a nullable time.Time to pgtype.Date, returning an error on failure.
func TryConvertToPgtypeDate(d *time.Time) (pgtype.Date, error) {
	if d == nil {
		return pgtype.Date{Valid: false}, nil
	}
	var pgDate pgtype.Date
	if err := pgDate.Scan(*d); err != nil {
		return pgtype.Date{Valid: false}, err
	}
	return pgDate, nil
}

// TryConvertToPgtypeTime converts a nullable time.Time to pgtype.Time, returning an error on failure.
func TryConvertToPgtypeTime(t *time.Time) (pgtype.Time, error) {
	if t == nil {
		return pgtype.Time{Valid: false}, nil
	}
	// pgtype.Time stores microseconds since midnight
	usec := int64(t.Hour())*3600_000_000 +
		int64(t.Minute())*60_000_000 +
		int64(t.Second())*1_000_000 +
		int64(t.Nanosecond())/1000
	return pgtype.Time{Microseconds: usec, Valid: true}, nil
}

// ConvertTimeOfDayToPgtypeTime converts a nullable rastime.TimeOfDay to pgtype.Time. Returns invalid if nil.
func ConvertTimeOfDayToPgtypeTime(t *rastime.TimeOfDay) pgtype.Time {
	if t == nil {
		return pgtype.Time{Valid: false}
	}
	usec := int64(t.Hour)*3600_000_000 + int64(t.Minute)*60_000_000
	return pgtype.Time{Microseconds: usec, Valid: true}
}

// ConvertPgtypeTimeToTimeOfDay converts a pgtype.Time to rastime.TimeOfDay. Returns nil if invalid.
func ConvertPgtypeTimeToTimeOfDay(t pgtype.Time) *rastime.TimeOfDay {
	if !t.Valid {
		return nil
	}
	totalMinutes := t.Microseconds / 60_000_000
	return &rastime.TimeOfDay{
		Hour:   int(totalMinutes / 60),
		Minute: int(totalMinutes % 60),
	}
}
