package config

import (
	"os"
	"testing"
	"time"
)

func TestGetEnvironmentVariableOrDefault(t *testing.T) {
	t.Run("returns value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")
		defer os.Unsetenv("TEST_VAR")

		result := GetEnvironmentVariableOrDefault("TEST_VAR", "default")
		if result != "test_value" {
			t.Errorf("expected test_value, got %s", result)
		}
	})

	t.Run("returns default when env var is not set", func(t *testing.T) {
		os.Unsetenv("TEST_VAR_NOT_SET")

		result := GetEnvironmentVariableOrDefault("TEST_VAR_NOT_SET", "default_value")
		if result != "default_value" {
			t.Errorf("expected default_value, got %s", result)
		}
	})

	t.Run("returns default when env var is empty", func(t *testing.T) {
		os.Setenv("TEST_VAR_EMPTY", "")
		defer os.Unsetenv("TEST_VAR_EMPTY")

		result := GetEnvironmentVariableOrDefault("TEST_VAR_EMPTY", "default_value")
		if result != "default_value" {
			t.Errorf("expected default_value, got %s", result)
		}
	})
}

func TestGetEnvironmentVariableOrPanic(t *testing.T) {
	t.Run("returns value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR_PANIC", "test_value")
		defer os.Unsetenv("TEST_VAR_PANIC")

		result := GetEnvironmentVariableOrPanic("TEST_VAR_PANIC", "should not panic")
		if result != "test_value" {
			t.Errorf("expected test_value, got %s", result)
		}
	})

	t.Run("panics when env var is not set", func(t *testing.T) {
		os.Unsetenv("TEST_VAR_PANIC_NOT_SET")

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when env var is not set")
			}
		}()

		GetEnvironmentVariableOrPanic("TEST_VAR_PANIC_NOT_SET", "panic message")
	})

	t.Run("panics when env var is empty", func(t *testing.T) {
		os.Setenv("TEST_VAR_PANIC_EMPTY", "")
		defer os.Unsetenv("TEST_VAR_PANIC_EMPTY")

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic when env var is empty")
			}
		}()

		GetEnvironmentVariableOrPanic("TEST_VAR_PANIC_EMPTY", "panic message")
	})
}

func TestGetEnvironmentVariableOrDefaultInt(t *testing.T) {
	t.Run("returns int value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_INT_VAR", "42")
		defer os.Unsetenv("TEST_INT_VAR")

		result := GetEnvironmentVariableOrDefaultInt("TEST_INT_VAR", 10)
		if result != 42 {
			t.Errorf("expected 42, got %d", result)
		}
	})

	t.Run("returns default when env var is not set", func(t *testing.T) {
		os.Unsetenv("TEST_INT_VAR_NOT_SET")

		result := GetEnvironmentVariableOrDefaultInt("TEST_INT_VAR_NOT_SET", 10)
		if result != 10 {
			t.Errorf("expected 10, got %d", result)
		}
	})

	t.Run("returns default when env var is not a valid int", func(t *testing.T) {
		os.Setenv("TEST_INT_VAR_INVALID", "not_a_number")
		defer os.Unsetenv("TEST_INT_VAR_INVALID")

		result := GetEnvironmentVariableOrDefaultInt("TEST_INT_VAR_INVALID", 10)
		if result != 10 {
			t.Errorf("expected 10, got %d", result)
		}
	})

	t.Run("handles negative numbers", func(t *testing.T) {
		os.Setenv("TEST_INT_VAR_NEG", "-5")
		defer os.Unsetenv("TEST_INT_VAR_NEG")

		result := GetEnvironmentVariableOrDefaultInt("TEST_INT_VAR_NEG", 10)
		if result != -5 {
			t.Errorf("expected -5, got %d", result)
		}
	})

	t.Run("handles zero", func(t *testing.T) {
		os.Setenv("TEST_INT_VAR_ZERO", "0")
		defer os.Unsetenv("TEST_INT_VAR_ZERO")

		result := GetEnvironmentVariableOrDefaultInt("TEST_INT_VAR_ZERO", 10)
		if result != 0 {
			t.Errorf("expected 0, got %d", result)
		}
	})
}

func TestGetEnvironmentVariableOrDefaultDuration(t *testing.T) {
	t.Run("returns duration when env var is set", func(t *testing.T) {
		os.Setenv("TEST_DUR_VAR", "5s")
		defer os.Unsetenv("TEST_DUR_VAR")

		result := GetEnvironmentVariableOrDefaultDuration("TEST_DUR_VAR", "1s")
		if result != 5*time.Second {
			t.Errorf("expected 5s, got %v", result)
		}
	})

	t.Run("returns default when env var is not set", func(t *testing.T) {
		os.Unsetenv("TEST_DUR_VAR_NOT_SET")

		result := GetEnvironmentVariableOrDefaultDuration("TEST_DUR_VAR_NOT_SET", "30s")
		if result != 30*time.Second {
			t.Errorf("expected 30s, got %v", result)
		}
	})

	t.Run("returns default when env var is invalid duration", func(t *testing.T) {
		os.Setenv("TEST_DUR_VAR_INVALID", "not_a_duration")
		defer os.Unsetenv("TEST_DUR_VAR_INVALID")

		result := GetEnvironmentVariableOrDefaultDuration("TEST_DUR_VAR_INVALID", "1m")
		if result != time.Minute {
			t.Errorf("expected 1m, got %v", result)
		}
	})

	t.Run("handles various duration formats", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected time.Duration
		}{
			{"100ms", 100 * time.Millisecond},
			{"2h", 2 * time.Hour},
			{"1h30m", 90 * time.Minute},
		}

		for _, tc := range testCases {
			os.Setenv("TEST_DUR_VAR_FMT", tc.input)
			result := GetEnvironmentVariableOrDefaultDuration("TEST_DUR_VAR_FMT", "1s")
			if result != tc.expected {
				t.Errorf("for input %s, expected %v, got %v", tc.input, tc.expected, result)
			}
		}
		os.Unsetenv("TEST_DUR_VAR_FMT")
	})
}
