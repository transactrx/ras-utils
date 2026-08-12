// Package rasvalidation provides common validation helpers for the Clinical+ ecosystem.
package rasvalidation

import (
	"net/mail"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/transactrx/ras-utils/rasformat"
)

var (
	// US phone: 10 digits, optional leading 1, balanced parens, common separators
	phoneRegex = regexp.MustCompile(`^1?[\s.-]?(\(\d{3}\)|\d{3})[\s.-]?\d{3}[\s.-]?\d{4}$`)

	// US ZIP: 5 digits or 5+4 format
	zipRegex = regexp.MustCompile(`^\d{5}(-\d{4})?$`)
)

// IsValidUUID returns true if s is a valid, non-nil UUID.
func IsValidUUID(s string) bool {
	if s == "" {
		return false
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return false
	}
	return id != uuid.Nil
}

// IsValidEmail returns true if s is a valid email address per RFC 5322.
func IsValidEmail(s string) bool {
	if s == "" {
		return false
	}
	_, err := mail.ParseAddress(s)
	return err == nil
}

// IsValidUSPhone returns true if s is a valid US phone number (10 digits).
func IsValidUSPhone(s string) bool {
	if s == "" {
		return false
	}
	return phoneRegex.MatchString(s)
}

// IsValidUSZip returns true if s is a valid US ZIP code (5 or 5+4 format).
func IsValidUSZip(s string) bool {
	if s == "" {
		return false
	}
	return zipRegex.MatchString(s)
}

// IsValidNPIFormat returns true if s is a valid NPI format (exactly 10 digits).
func IsValidNPIFormat(s string) bool {
	if s == "" || len(s) != 10 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// IsValidNPIChecksum returns true if s passes the NPI Luhn checksum validation.
// Assumes s is already a valid 10-digit format.
func IsValidNPIChecksum(s string) bool {
	if len(s) != 10 {
		return false
	}
	// Luhn check with 80840 prefix per CMS standard
	prefixed := "80840" + s
	sum := 0
	for i := len(prefixed) - 1; i >= 0; i-- {
		d := int(prefixed[i] - '0')
		if (len(prefixed)-1-i)%2 == 1 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return sum%10 == 0
}

// IsValidNPI returns true if s is a valid NPI (10 digits with valid Luhn checksum).
func IsValidNPI(s string) bool {
	return IsValidNPIFormat(s) && IsValidNPIChecksum(s)
}

// IsValidDateString returns true if s can be parsed with the given layout.
func IsValidDateString(s string, layout string) bool {
	if s == "" {
		return false
	}
	_, err := time.Parse(layout, s)
	return err == nil
}

// IsValidISO8601Date returns true if s is a valid ISO 8601 date (YYYY-MM-DD).
func IsValidISO8601Date(s string) bool {
	return IsValidDateString(s, "2006-01-02")
}

// IsValidMMDDYYYYDate returns true if s is a valid MM/DD/YYYY date.
func IsValidMMDDYYYYDate(s string) bool {
	return IsValidDateString(s, "01/02/2006")
}


// MaskPhone masks a phone number for logging, showing only the last 4 digits.
// Returns "***-***-XXXX" format for valid 10-digit numbers, or "****" for others.
//
// Deprecated: Use rasformat.MaskPhone instead.
func MaskPhone(phone string) string {
	return rasformat.MaskPhone(phone)
}

// MaskEmail masks an email address for logging, showing first char and domain.
// Returns "j***@example.com" format, or "****" if invalid.
//
// Deprecated: Use rasformat.MaskEmail instead.
func MaskEmail(email string) string {
	return rasformat.MaskEmail(email)
}
