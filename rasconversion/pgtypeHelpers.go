package rasconversion

import (
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

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

// Error-returning variants for callers that need explicit error handling

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

func TryConvertToPgtypeInt2(i *int32) (pgtype.Int2, error) {
	if i == nil {
		return pgtype.Int2{Valid: false}, nil
	}
	return pgtype.Int2{
		Int16: int16(*i),
		Valid: true,
	}, nil
}

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
