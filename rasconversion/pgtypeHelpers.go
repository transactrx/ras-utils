// Package rasconversion provides type conversion helpers for PostgreSQL using pgx/pgtype.
//
// It converts nullable Go types (*string, *int64, *time.Time, etc.) to their pgtype
// equivalents (pgtype.Text, pgtype.Int8, pgtype.Timestamp, etc.) with proper null handling.
//
// Two variants are provided for each conversion:
//   - ConvertTo* functions log errors and return invalid pgtypes on failure
//   - TryConvertTo* functions return errors for explicit error handling
//
// Reverse conversions (pgtype to Go pointer) are also provided:
//   - ConvertFrom* functions return nil for invalid pgtypes
package rasconversion

import (
	"encoding/json"
	"log/slog"
	"math/big"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/transactrx/ras-utils/rastime"
)

// ConvertToPgtypeFloat4 converts a nullable float32 to pgtype.Float4. Returns invalid if nil.
func ConvertToPgtypeFloat4(f *float32) pgtype.Float4 {
	if f != nil {
		return pgtype.Float4{Float32: *f, Valid: true}
	}
	return pgtype.Float4{Valid: false}
}

// ConvertToPgtypeInterval converts a nullable time.Duration to pgtype.Interval. Returns invalid if nil.
func ConvertToPgtypeInterval(d *time.Duration) pgtype.Interval {
	if d == nil {
		return pgtype.Interval{Valid: false}
	}
	return pgtype.Interval{Microseconds: d.Microseconds(), Valid: true}
}

// ConvertToPgtypeJSONB converts any value to a JSONB-compatible byte slice. Returns nil on error.
func ConvertToPgtypeJSONB(v any) []byte {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		slog.Error("failed to marshal to JSONB", "error", err)
		return nil
	}
	return b
}

// TryConvertToPgtypeJSONB converts any value to a JSONB-compatible byte slice, returning an error on failure.
func TryConvertToPgtypeJSONB(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}

// ConvertToPgtypeInt4 converts a nullable int32 to pgtype.Int4. Returns invalid if nil.
func ConvertToPgtypeInt4(i *int32) pgtype.Int4 {
	if i != nil {
		return pgtype.Int4{Int32: *i, Valid: true}
	}
	return pgtype.Int4{Valid: false}
}

// ConvertToPgtypeFloat8 converts a nullable float64 to pgtype.Float8. Returns invalid if nil.
func ConvertToPgtypeFloat8(f *float64) pgtype.Float8 {
	if f != nil {
		return pgtype.Float8{Float64: *f, Valid: true}
	}
	return pgtype.Float8{Valid: false}
}

// ConvertToPgtypeNumeric converts a nullable float64 to pgtype.Numeric. Returns invalid if nil.
func ConvertToPgtypeNumeric(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	var n pgtype.Numeric
	if err := n.Scan(strconv.FormatFloat(*f, 'f', -1, 64)); err != nil {
		slog.Error("failed to convert to pgtype.Numeric", "error", err, "float", *f)
		return pgtype.Numeric{Valid: false}
	}
	return n
}

// ConvertToPgtypeUUID converts a nullable uuid.UUID to pgtype.UUID. Returns invalid if nil.
func ConvertToPgtypeUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *u, Valid: true}
}

// ConvertToPgtypeUUIDFromString converts a nullable string to pgtype.UUID. Returns invalid if nil or invalid UUID.
func ConvertToPgtypeUUIDFromString(s *string) pgtype.UUID {
	if s == nil {
		return pgtype.UUID{Valid: false}
	}
	u, err := uuid.Parse(*s)
	if err != nil {
		slog.Error("failed to parse UUID string", "error", err, "string", *s)
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: u, Valid: true}
}

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

// TryConvertToPgtypeInt4 converts a nullable int32 to pgtype.Int4, returning an error on failure.
func TryConvertToPgtypeInt4(i *int32) (pgtype.Int4, error) {
	if i == nil {
		return pgtype.Int4{Valid: false}, nil
	}
	return pgtype.Int4{Int32: *i, Valid: true}, nil
}

// TryConvertToPgtypeFloat4 converts a nullable float32 to pgtype.Float4, returning an error on failure.
func TryConvertToPgtypeFloat4(f *float32) (pgtype.Float4, error) {
	if f == nil {
		return pgtype.Float4{Valid: false}, nil
	}
	return pgtype.Float4{Float32: *f, Valid: true}, nil
}

// TryConvertToPgtypeInterval converts a nullable time.Duration to pgtype.Interval, returning an error on failure.
func TryConvertToPgtypeInterval(d *time.Duration) (pgtype.Interval, error) {
	if d == nil {
		return pgtype.Interval{Valid: false}, nil
	}
	return pgtype.Interval{Microseconds: d.Microseconds(), Valid: true}, nil
}

// TryConvertToPgtypeFloat8 converts a nullable float64 to pgtype.Float8, returning an error on failure.
func TryConvertToPgtypeFloat8(f *float64) (pgtype.Float8, error) {
	if f == nil {
		return pgtype.Float8{Valid: false}, nil
	}
	return pgtype.Float8{Float64: *f, Valid: true}, nil
}

// TryConvertToPgtypeNumeric converts a nullable float64 to pgtype.Numeric, returning an error on failure.
func TryConvertToPgtypeNumeric(f *float64) (pgtype.Numeric, error) {
	if f == nil {
		return pgtype.Numeric{Valid: false}, nil
	}
	var n pgtype.Numeric
	if err := n.Scan(strconv.FormatFloat(*f, 'f', -1, 64)); err != nil {
		return pgtype.Numeric{Valid: false}, err
	}
	return n, nil
}

// TryConvertToPgtypeUUID converts a nullable uuid.UUID to pgtype.UUID, returning an error on failure.
func TryConvertToPgtypeUUID(u *uuid.UUID) (pgtype.UUID, error) {
	if u == nil {
		return pgtype.UUID{Valid: false}, nil
	}
	return pgtype.UUID{Bytes: *u, Valid: true}, nil
}

// TryConvertToPgtypeUUIDFromString converts a nullable string to pgtype.UUID, returning an error on failure.
func TryConvertToPgtypeUUIDFromString(s *string) (pgtype.UUID, error) {
	if s == nil {
		return pgtype.UUID{Valid: false}, nil
	}
	u, err := uuid.Parse(*s)
	if err != nil {
		return pgtype.UUID{Valid: false}, err
	}
	return pgtype.UUID{Bytes: u, Valid: true}, nil
}

// Reverse conversions: pgtype to Go pointer

// ConvertFromPgtypeText converts pgtype.Text to *string. Returns nil if invalid.
func ConvertFromPgtypeText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// ConvertFromPgtypeInt2 converts pgtype.Int2 to *int16. Returns nil if invalid.
func ConvertFromPgtypeInt2(i pgtype.Int2) *int16 {
	if !i.Valid {
		return nil
	}
	return &i.Int16
}

// ConvertFromPgtypeInt4 converts pgtype.Int4 to *int32. Returns nil if invalid.
func ConvertFromPgtypeInt4(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return &i.Int32
}

// ConvertFromPgtypeInt8 converts pgtype.Int8 to *int64. Returns nil if invalid.
func ConvertFromPgtypeInt8(i pgtype.Int8) *int64 {
	if !i.Valid {
		return nil
	}
	return &i.Int64
}

// ConvertFromPgtypeBool converts pgtype.Bool to *bool. Returns nil if invalid.
func ConvertFromPgtypeBool(b pgtype.Bool) *bool {
	if !b.Valid {
		return nil
	}
	return &b.Bool
}

// ConvertFromPgtypeFloat8 converts pgtype.Float8 to *float64. Returns nil if invalid.
func ConvertFromPgtypeFloat8(f pgtype.Float8) *float64 {
	if !f.Valid {
		return nil
	}
	return &f.Float64
}

// ConvertFromPgtypeNumeric converts pgtype.Numeric to *float64. Returns nil if invalid.
func ConvertFromPgtypeNumeric(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f, _ := n.Float64Value()
	if !f.Valid {
		return nil
	}
	return &f.Float64
}

// ConvertFromPgtypeNumericToBigFloat converts pgtype.Numeric to *big.Float for precision-sensitive operations.
func ConvertFromPgtypeNumericToBigFloat(n pgtype.Numeric) *big.Float {
	if !n.Valid {
		return nil
	}
	bf := new(big.Float)
	if n.Int != nil {
		bf.SetInt(n.Int)
	}
	if n.Exp != 0 {
		exp := new(big.Float).SetInt(big.NewInt(10))
		exp = exp.SetMantExp(exp, int(n.Exp))
		bf.Mul(bf, exp)
	}
	return bf
}

// ConvertFromPgtypeTimestamp converts pgtype.Timestamp to *time.Time. Returns nil if invalid.
func ConvertFromPgtypeTimestamp(t pgtype.Timestamp) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// ConvertFromPgtypeTimestamptz converts pgtype.Timestamptz to *time.Time. Returns nil if invalid.
func ConvertFromPgtypeTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// ConvertFromPgtypeDate converts pgtype.Date to *time.Time. Returns nil if invalid.
func ConvertFromPgtypeDate(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	return &d.Time
}

// ConvertFromPgtypeTime converts pgtype.Time to *time.Time. Returns nil if invalid.
func ConvertFromPgtypeTime(t pgtype.Time) *time.Time {
	if !t.Valid {
		return nil
	}
	hours := t.Microseconds / 3600_000_000
	remaining := t.Microseconds % 3600_000_000
	minutes := remaining / 60_000_000
	remaining = remaining % 60_000_000
	seconds := remaining / 1_000_000
	micros := remaining % 1_000_000
	tm := time.Date(0, 1, 1, int(hours), int(minutes), int(seconds), int(micros)*1000, time.UTC)
	return &tm
}

// ConvertFromPgtypeUUID converts pgtype.UUID to *uuid.UUID. Returns nil if invalid.
func ConvertFromPgtypeUUID(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	result := uuid.UUID(u.Bytes)
	return &result
}

// ConvertFromPgtypeUUIDToString converts pgtype.UUID to *string. Returns nil if invalid.
func ConvertFromPgtypeUUIDToString(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := uuid.UUID(u.Bytes).String()
	return &s
}

// ConvertFromPgtypeFloat4 converts pgtype.Float4 to *float32. Returns nil if invalid.
func ConvertFromPgtypeFloat4(f pgtype.Float4) *float32 {
	if !f.Valid {
		return nil
	}
	return &f.Float32
}

// ConvertFromPgtypeInterval converts pgtype.Interval to *time.Duration. Returns nil if invalid.
func ConvertFromPgtypeInterval(i pgtype.Interval) *time.Duration {
	if !i.Valid {
		return nil
	}
	d := time.Duration(i.Microseconds) * time.Microsecond
	if i.Days != 0 {
		d += time.Duration(i.Days) * 24 * time.Hour
	}
	if i.Months != 0 {
		d += time.Duration(i.Months) * 30 * 24 * time.Hour
	}
	return &d
}

// ConvertFromPgtypeJSONB unmarshals JSONB bytes into the provided type. Returns nil on error or nil input.
func ConvertFromPgtypeJSONB[T any](b []byte) *T {
	if b == nil {
		return nil
	}
	var result T
	if err := json.Unmarshal(b, &result); err != nil {
		slog.Error("failed to unmarshal JSONB", "error", err)
		return nil
	}
	return &result
}

// TryConvertFromPgtypeJSONB unmarshals JSONB bytes into the provided type, returning an error on failure.
func TryConvertFromPgtypeJSONB[T any](b []byte) (*T, error) {
	if b == nil {
		return nil, nil
	}
	var result T
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// OrDefault variants: return default value instead of nil for invalid pgtypes

// ConvertFromPgtypeTextOrDefault converts pgtype.Text to string, returning def if invalid.
func ConvertFromPgtypeTextOrDefault(t pgtype.Text, def string) string {
	if !t.Valid {
		return def
	}
	return t.String
}

// ConvertFromPgtypeInt2OrDefault converts pgtype.Int2 to int16, returning def if invalid.
func ConvertFromPgtypeInt2OrDefault(i pgtype.Int2, def int16) int16 {
	if !i.Valid {
		return def
	}
	return i.Int16
}

// ConvertFromPgtypeInt4OrDefault converts pgtype.Int4 to int32, returning def if invalid.
func ConvertFromPgtypeInt4OrDefault(i pgtype.Int4, def int32) int32 {
	if !i.Valid {
		return def
	}
	return i.Int32
}

// ConvertFromPgtypeInt8OrDefault converts pgtype.Int8 to int64, returning def if invalid.
func ConvertFromPgtypeInt8OrDefault(i pgtype.Int8, def int64) int64 {
	if !i.Valid {
		return def
	}
	return i.Int64
}

// ConvertFromPgtypeBoolOrDefault converts pgtype.Bool to bool, returning def if invalid.
func ConvertFromPgtypeBoolOrDefault(b pgtype.Bool, def bool) bool {
	if !b.Valid {
		return def
	}
	return b.Bool
}

// ConvertFromPgtypeFloat4OrDefault converts pgtype.Float4 to float32, returning def if invalid.
func ConvertFromPgtypeFloat4OrDefault(f pgtype.Float4, def float32) float32 {
	if !f.Valid {
		return def
	}
	return f.Float32
}

// ConvertFromPgtypeFloat8OrDefault converts pgtype.Float8 to float64, returning def if invalid.
func ConvertFromPgtypeFloat8OrDefault(f pgtype.Float8, def float64) float64 {
	if !f.Valid {
		return def
	}
	return f.Float64
}

// ConvertFromPgtypeNumericOrDefault converts pgtype.Numeric to float64, returning def if invalid.
func ConvertFromPgtypeNumericOrDefault(n pgtype.Numeric, def float64) float64 {
	if !n.Valid {
		return def
	}
	f, _ := n.Float64Value()
	if !f.Valid {
		return def
	}
	return f.Float64
}

// ConvertFromPgtypeTimestampOrDefault converts pgtype.Timestamp to time.Time, returning def if invalid.
func ConvertFromPgtypeTimestampOrDefault(t pgtype.Timestamp, def time.Time) time.Time {
	if !t.Valid {
		return def
	}
	return t.Time
}

// ConvertFromPgtypeTimestamptzOrDefault converts pgtype.Timestamptz to time.Time, returning def if invalid.
func ConvertFromPgtypeTimestamptzOrDefault(t pgtype.Timestamptz, def time.Time) time.Time {
	if !t.Valid {
		return def
	}
	return t.Time
}

// ConvertFromPgtypeDateOrDefault converts pgtype.Date to time.Time, returning def if invalid.
func ConvertFromPgtypeDateOrDefault(d pgtype.Date, def time.Time) time.Time {
	if !d.Valid {
		return def
	}
	return d.Time
}

// ConvertFromPgtypeIntervalOrDefault converts pgtype.Interval to time.Duration, returning def if invalid.
func ConvertFromPgtypeIntervalOrDefault(i pgtype.Interval, def time.Duration) time.Duration {
	if !i.Valid {
		return def
	}
	d := time.Duration(i.Microseconds) * time.Microsecond
	if i.Days != 0 {
		d += time.Duration(i.Days) * 24 * time.Hour
	}
	if i.Months != 0 {
		d += time.Duration(i.Months) * 30 * 24 * time.Hour
	}
	return d
}

// ConvertFromPgtypeUUIDOrDefault converts pgtype.UUID to uuid.UUID, returning def if invalid.
func ConvertFromPgtypeUUIDOrDefault(u pgtype.UUID, def uuid.UUID) uuid.UUID {
	if !u.Valid {
		return def
	}
	return uuid.UUID(u.Bytes)
}

// ConvertFromPgtypeUUIDToStringOrDefault converts pgtype.UUID to string, returning def if invalid.
func ConvertFromPgtypeUUIDToStringOrDefault(u pgtype.UUID, def string) string {
	if !u.Valid {
		return def
	}
	return uuid.UUID(u.Bytes).String()
}
