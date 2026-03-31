package conversion

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
			slog.Error("failed to convert to pgtype.Bool", "error", err, "bool", t)
			pgTime = pgtype.Timestamp{Time: *t}
		}
	} else {
		pgTime = pgtype.Timestamp{Time: *t}
	}
	return pgTime
}
