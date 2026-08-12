package rasformat

import (
	"strings"
	"time"
)

// NormalizePhone strips all non-digit characters from a phone number.
// Removes leading "1" country code if present, returning 10 digits.
func NormalizePhone(phone string) string {
	digits := make([]byte, 0, len(phone))
	for i := 0; i < len(phone); i++ {
		if phone[i] >= '0' && phone[i] <= '9' {
			digits = append(digits, phone[i])
		}
	}

	if len(digits) == 11 && digits[0] == '1' {
		digits = digits[1:]
	}

	return string(digits)
}

// FormatPhoneParens formats a phone number as (XXX) XXX-XXXX.
// Returns empty string if input doesn't contain exactly 10 digits.
func FormatPhoneParens(phone string) string {
	digits := NormalizePhone(phone)
	if len(digits) != 10 {
		return ""
	}
	return "(" + digits[:3] + ") " + digits[3:6] + "-" + digits[6:]
}

// FormatPhoneDots formats a phone number as XXX.XXX.XXXX.
// Returns empty string if input doesn't contain exactly 10 digits.
func FormatPhoneDots(phone string) string {
	digits := NormalizePhone(phone)
	if len(digits) != 10 {
		return ""
	}
	return digits[:3] + "." + digits[3:6] + "." + digits[6:]
}

// FormatPhoneDashes formats a phone number as XXX-XXX-XXXX.
// Returns empty string if input doesn't contain exactly 10 digits.
func FormatPhoneDashes(phone string) string {
	digits := NormalizePhone(phone)
	if len(digits) != 10 {
		return ""
	}
	return digits[:3] + "-" + digits[3:6] + "-" + digits[6:]
}

// NormalizeDateISO parses a date string in various formats and returns ISO 8601 (YYYY-MM-DD).
// Supported formats: RFC3339, ISO date, compact (YYYYMMDD).
// Returns error if date cannot be parsed.
func NormalizeDateISO(date string) (string, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"20060102",
		"01/02/2006",
		"1/2/2006",
	}

	for _, format := range formats {
		t, err := time.Parse(format, date)
		if err == nil {
			return t.Format("2006-01-02"), nil
		}
	}

	return "", &ParseError{Input: date, Kind: "date"}
}

// MaskPhone masks a phone number for logging, showing only the last 4 digits.
// Returns "***-***-XXXX" format for valid numbers with 4+ digits, or "****" for others.
func MaskPhone(phone string) string {
	digits := NormalizePhone(phone)

	if len(digits) >= 4 {
		return "***-***-" + digits[len(digits)-4:]
	}
	return "****"
}

// MaskEmail masks an email address for logging, showing first char and domain.
// Returns "j***@example.com" format, or "****" if invalid.
func MaskEmail(email string) string {
	if email == "" {
		return "****"
	}

	atIdx := strings.IndexByte(email, '@')
	if atIdx <= 0 {
		return "****"
	}

	return string(email[0]) + "***" + email[atIdx:]
}

// MaskName masks a name for logging, showing only the first character.
// Returns "J***" format, or "****" if empty.
func MaskName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "****"
	}
	return string(name[0]) + "***"
}

// MaskFullName masks a full name (first + last) for logging.
// Returns "J*** S***" format.
func MaskFullName(firstName, lastName string) string {
	return MaskName(firstName) + " " + MaskName(lastName)
}

// MaskDOB masks a date of birth for logging, showing only the day.
// Expects ISO format (YYYY-MM-DD). Returns "****-**-DD" or "****" if invalid.
func MaskDOB(dob string) string {
	if len(dob) < 10 {
		return "****"
	}
	// Check for ISO format: YYYY-MM-DD
	if dob[4] == '-' && dob[7] == '-' {
		return "****-**-" + dob[8:10]
	}
	return "****"
}

// ParseError indicates a value could not be parsed.
type ParseError struct {
	Input string
	Kind  string
}

func (e *ParseError) Error() string {
	return "unable to parse " + e.Kind + ": " + e.Input
}
