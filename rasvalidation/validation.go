// Package rasvalidation provides common validation helpers for the Clinical+ ecosystem.
package rasvalidation

import (
	"net/mail"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	// US phone: 10 digits, optional leading 1, allows common separators
	phoneRegex = regexp.MustCompile(`^1?[\s.-]?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}$`)

	// US ZIP: 5 digits or 5+4 format
	zipRegex = regexp.MustCompile(`^\d{5}(-\d{4})?$`)

	// NPI: exactly 10 digits
	npiRegex = regexp.MustCompile(`^\d{10}$`)
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

// IsValidNPI returns true if s is a valid NPI (10 digits).
func IsValidNPI(s string) bool {
	if s == "" {
		return false
	}
	return npiRegex.MatchString(s)
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
