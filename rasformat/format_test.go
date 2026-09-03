package rasformat

import "testing"

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  string
	}{
		{"10 digit", "5551234567", "5551234567"},
		{"formatted parens", "(555) 123-4567", "5551234567"},
		{"with country code", "15551234567", "5551234567"},
		{"with dashes", "555-123-4567", "5551234567"},
		{"short", "123", "123"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizePhone(tt.phone); got != tt.want {
				t.Errorf("NormalizePhone(%q) = %q, want %q", tt.phone, got, tt.want)
			}
		})
	}
}

func TestFormatPhoneParens(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  string
	}{
		{"10 digit", "5551234567", "(555) 123-4567"},
		{"already formatted", "(555) 123-4567", "(555) 123-4567"},
		{"with country code", "15551234567", "(555) 123-4567"},
		{"short", "123", ""},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatPhoneParens(tt.phone); got != tt.want {
				t.Errorf("FormatPhoneParens(%q) = %q, want %q", tt.phone, got, tt.want)
			}
		})
	}
}

func TestFormatPhoneDots(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  string
	}{
		{"10 digit", "5551234567", "555.123.4567"},
		{"formatted", "(555) 123-4567", "555.123.4567"},
		{"short", "123", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatPhoneDots(tt.phone); got != tt.want {
				t.Errorf("FormatPhoneDots(%q) = %q, want %q", tt.phone, got, tt.want)
			}
		})
	}
}

func TestFormatPhoneDashes(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  string
	}{
		{"10 digit", "5551234567", "555-123-4567"},
		{"formatted", "(555) 123-4567", "555-123-4567"},
		{"short", "123", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatPhoneDashes(tt.phone); got != tt.want {
				t.Errorf("FormatPhoneDashes(%q) = %q, want %q", tt.phone, got, tt.want)
			}
		})
	}
}

func TestNormalizeDateISO(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		want    string
		wantErr bool
	}{
		{"ISO format", "2024-03-15", "2024-03-15", false},
		{"RFC3339", "2024-03-15T10:30:00Z", "2024-03-15", false},
		{"compact", "20240315", "2024-03-15", false},
		{"US format", "03/15/2024", "2024-03-15", false},
		{"US format no padding", "3/5/2024", "2024-03-05", false},
		{"invalid", "not-a-date", "", true},
		{"empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeDateISO(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeDateISO(%q) error = %v, wantErr %v", tt.date, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NormalizeDateISO(%q) = %q, want %q", tt.date, got, tt.want)
			}
		})
	}
}

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  string
	}{
		{"10 digit", "5551234567", "***-***-4567"},
		{"formatted", "(555) 123-4567", "***-***-4567"},
		{"with country code", "15551234567", "***-***-4567"},
		{"short", "123", "****"},
		{"empty", "", "****"},
		{"with dashes", "555-123-4567", "***-***-4567"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskPhone(tt.phone); got != tt.want {
				t.Errorf("MaskPhone(%q) = %q, want %q", tt.phone, got, tt.want)
			}
		})
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{"standard", "john@example.com", "j***@example.com"},
		{"single char local", "j@example.com", "j***@example.com"},
		{"empty", "", "****"},
		{"no at", "invalid", "****"},
		{"at start", "@example.com", "****"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskEmail(tt.email); got != tt.want {
				t.Errorf("MaskEmail(%q) = %q, want %q", tt.email, got, tt.want)
			}
		})
	}
}

func TestMaskName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"standard", "John", "J***"},
		{"lowercase", "john", "j***"},
		{"single char", "J", "J***"},
		{"empty", "", "****"},
		{"whitespace", "  ", "****"},
		{"with spaces", "  John  ", "J***"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskName(tt.in); got != tt.want {
				t.Errorf("MaskName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestMaskFullName(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		want      string
	}{
		{"standard", "John", "Smith", "J*** S***"},
		{"empty first", "", "Smith", "**** S***"},
		{"empty last", "John", "", "J*** ****"},
		{"both empty", "", "", "**** ****"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskFullName(tt.firstName, tt.lastName); got != tt.want {
				t.Errorf("MaskFullName(%q, %q) = %q, want %q", tt.firstName, tt.lastName, got, tt.want)
			}
		})
	}
}

func TestMaskDOB(t *testing.T) {
	tests := []struct {
		name string
		dob  string
		want string
	}{
		{"ISO format", "1990-05-15", "****-**-15"},
		{"with time", "1990-05-15T00:00:00Z", "****-**-15"},
		{"short", "1990", "****"},
		{"empty", "", "****"},
		{"invalid format", "05/15/1990", "****"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskDOB(tt.dob); got != tt.want {
				t.Errorf("MaskDOB(%q) = %q, want %q", tt.dob, got, tt.want)
			}
		})
	}
}

func TestMaskWords(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		visibleChars int
		mask         rune
		want         string
	}{
		{"single word", "Hello", 1, '*', "H****"},
		{"two words", "John Smith", 1, '*', "J*** S****"},
		{"show two chars", "Hello World", 2, '*', "He*** Wo***"},
		{"custom mask", "Secret Data", 1, '#', "S##### D###"},
		{"empty", "", 1, '*', "****"},
		{"whitespace only", "   ", 1, '*', "****"},
		{"word shorter than visible", "Hi", 3, '*', "Hi"},
		{"word equal to visible", "Cat", 3, '*', "Cat"},
		{"zero visible", "Hello", 0, '*', "*****"},
		{"negative visible", "Hello", -1, '*', "*****"},
		{"multiple spaces", "John    Smith", 1, '*', "J*** S****"},
		{"unicode", "José María", 1, '*', "J*** M****"},
		{"unicode mask", "Hello", 1, '█', "H████"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskWords(tt.value, tt.visibleChars, tt.mask); got != tt.want {
				t.Errorf("MaskWords(%q, %d, %q) = %q, want %q", tt.value, tt.visibleChars, tt.mask, got, tt.want)
			}
		})
	}
}

func TestMaskWordsFixed(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		visibleChars int
		mask         rune
		maskLength   int
		want         string
	}{
		{"single word", "Hello", 1, '*', 3, "H***"},
		{"two words", "John Smith", 1, '*', 3, "J*** S***"},
		{"long mask", "Hi There", 1, '*', 5, "H***** T*****"},
		{"short mask", "Hello World", 2, '*', 1, "He* Wo*"},
		{"empty", "", 1, '*', 3, "***"},
		{"whitespace only", "   ", 1, '*', 3, "***"},
		{"word shorter than visible", "Hi", 3, '*', 3, "Hi"},
		{"word equal to visible", "Cat", 3, '*', 3, "Cat"},
		{"zero visible", "Hello", 0, '*', 4, "****"},
		{"custom mask", "Secret", 1, '#', 3, "S###"},
		{"unicode", "José", 1, '*', 3, "J***"},
		{"zero mask length falls back to word length", "Hello", 1, '*', 0, "H****"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskWordsFixed(tt.value, tt.visibleChars, tt.mask, tt.maskLength); got != tt.want {
				t.Errorf("MaskWordsFixed(%q, %d, %q, %d) = %q, want %q", tt.value, tt.visibleChars, tt.mask, tt.maskLength, got, tt.want)
			}
		})
	}
}
