package rasvalidation

import "testing"

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid uuid uppercase", "550E8400-E29B-41D4-A716-446655440000", true},
		{"valid uuid v4", "f47ac10b-58cc-4372-a567-0e02b2c3d479", true},
		{"nil uuid", "00000000-0000-0000-0000-000000000000", false},
		{"empty string", "", false},
		{"invalid format", "not-a-uuid", false},
		{"too short", "550e8400-e29b-41d4", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidUUID(tt.input); got != tt.want {
				t.Errorf("IsValidUUID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid email", "test@example.com", true},
		{"valid with subdomain", "test@mail.example.com", true},
		{"valid with plus", "test+tag@example.com", true},
		{"empty string", "", false},
		{"no at sign", "testexample.com", false},
		{"no domain", "test@", false},
		{"no local part", "@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidEmail(tt.input); got != tt.want {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidUSPhone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"10 digits plain", "5551234567", true},
		{"with dashes", "555-123-4567", true},
		{"with dots", "555.123.4567", true},
		{"with parens", "(555)123-4567", true},
		{"with parens and dashes", "(555)-123-4567", true},
		{"with leading 1", "15551234567", true},
		{"with leading 1 and dashes", "1-555-123-4567", true},
		{"with leading 1 and parens", "1-(555)-123-4567", true},
		{"empty string", "", false},
		{"too short", "555123456", false},
		{"too long", "55512345678", false},
		{"letters", "555-ABC-4567", false},
		{"unbalanced paren open only", "(555123-4567", false},
		{"unbalanced paren close only", "555)123-4567", false},
		{"misplaced space", "5 55-123-4567", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidUSPhone(tt.input); got != tt.want {
				t.Errorf("IsValidUSPhone(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidUSZip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"5 digit", "12345", true},
		{"5+4 format", "12345-6789", true},
		{"empty string", "", false},
		{"too short", "1234", false},
		{"too long", "123456", false},
		{"letters", "1234A", false},
		{"invalid plus4", "12345-678", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidUSZip(tt.input); got != tt.want {
				t.Errorf("IsValidUSZip(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidNPI(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid NPI", "1234567893", true},
		{"valid NPI 2", "1245319599", true},
		{"valid NPI 3", "1003000126", true},
		{"valid NPI 4", "1003968488", true},
		{"invalid checksum", "1234567890", false},
		{"all zeros", "0000000000", false},
		{"empty string", "", false},
		{"too short", "123456789", false},
		{"too long", "12345678901", false},
		{"letters", "123456789A", false},
		{"with dashes", "123-456-7890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidNPI(tt.input); got != tt.want {
				t.Errorf("IsValidNPI(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidISO8601Date(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid date", "2024-01-15", true},
		{"valid leap year", "2024-02-29", true},
		{"empty string", "", false},
		{"wrong format", "01/15/2024", false},
		{"invalid month", "2024-13-01", false},
		{"invalid day", "2024-01-32", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidISO8601Date(tt.input); got != tt.want {
				t.Errorf("IsValidISO8601Date(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidMMDDYYYYDate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid date", "01/15/2024", true},
		{"valid leap year", "02/29/2024", true},
		{"empty string", "", false},
		{"wrong format", "2024-01-15", false},
		{"invalid month", "13/01/2024", false},
		{"invalid day", "01/32/2024", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidMMDDYYYYDate(tt.input); got != tt.want {
				t.Errorf("IsValidMMDDYYYYDate(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}


func TestMaskPhone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"10 digits plain", "5551234567", "***-***-4567"},
		{"with dashes", "555-123-4567", "***-***-4567"},
		{"with dots", "555.123.4567", "***-***-4567"},
		{"with parens", "(555)123-4567", "***-***-4567"},
		{"with leading 1", "15551234567", "***-***-4567"},
		{"with leading 1 and dashes", "1-555-123-4567", "***-***-4567"},
		{"empty string", "", "****"},
		{"too short", "123", "****"},
		{"4 digits only", "4567", "***-***-4567"},
		{"7 digits", "1234567", "***-***-4567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskPhone(tt.input); got != tt.want {
				t.Errorf("MaskPhone(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple email", "john@example.com", "j***@example.com"},
		{"single char local", "j@example.com", "j***@example.com"},
		{"long local part", "johnathan.doe@example.com", "j***@example.com"},
		{"with subdomain", "test@mail.example.com", "t***@mail.example.com"},
		{"empty string", "", "****"},
		{"no at sign", "testexample.com", "****"},
		{"at sign at start", "@example.com", "****"},
		{"only at sign", "@", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskEmail(tt.input); got != tt.want {
				t.Errorf("MaskEmail(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
