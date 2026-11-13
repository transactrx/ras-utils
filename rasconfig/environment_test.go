package rasconfig

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

func TestTryGetEnvironmentVariable(t *testing.T) {
	t.Run("returns bool when env var is set to true", func(t *testing.T) {
		os.Setenv("TEST_VAR", "true")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[bool]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if !result {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns bool when env var is set to false", func(t *testing.T) {
		os.Setenv("TEST_VAR", "false")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[bool]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result {
			t.Error("expected false, got true")
		}
	})

	t.Run("returns int16 when env var is set to 69", func(t *testing.T) {
		os.Setenv("TEST_VAR", "69")
		defer os.Unsetenv("TEST_VAR")

		var zero int16

		result, err := TryGetEnvironmentVariable[int16]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result == zero {
			t.Errorf("expected 69, got %v", result)
		}
	})

	t.Run("returns int8 when env var is set to 69", func(t *testing.T) {
		os.Setenv("TEST_VAR", "69")
		defer os.Unsetenv("TEST_VAR")

		var zero int8

		result, err := TryGetEnvironmentVariable[int8]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result == zero {
			t.Errorf("expected 69, got %v", result)
		}
	})

	t.Run("returns int32 when env var is set to 69", func(t *testing.T) {
		os.Setenv("TEST_VAR", "69")
		defer os.Unsetenv("TEST_VAR")

		var zero int32

		result, err := TryGetEnvironmentVariable[int32]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result == zero {
			t.Errorf("expected 69, got %v", result)
		}
	})

	t.Run("returns int64 when env var is set to 69", func(t *testing.T) {
		os.Setenv("TEST_VAR", "69")
		defer os.Unsetenv("TEST_VAR")

		var zero int64

		result, err := TryGetEnvironmentVariable[int64]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result == zero {
			t.Errorf("expected 69, got %v", result)
		}
	})

	t.Run("returns float32 when env var is set to 69.69", func(t *testing.T) {
		os.Setenv("TEST_VAR", "69.69")
		defer os.Unsetenv("TEST_VAR")

		var zero float32

		result, err := TryGetEnvironmentVariable[float32]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result == zero {
			t.Errorf("expected 69, got %v", result)
		}
	})

	t.Run("returns float64 when env var is set to 69.69", func(t *testing.T) {
		os.Setenv("TEST_VAR", "69.69")
		defer os.Unsetenv("TEST_VAR")

		var zero float64

		result, err := TryGetEnvironmentVariable[float64]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result == zero {
			t.Errorf("expected 69, got %v", result)
		}
	})

	t.Run("returns float64 when env var is set to 69.69", func(t *testing.T) {
		os.Setenv("TEST_VAR", "69.69")
		defer os.Unsetenv("TEST_VAR")

		var zero float64

		result, err := TryGetEnvironmentVariable[float64]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result == zero {
			t.Errorf("expected 69, got %v", result)
		}
	})

	t.Run("returns string when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "hello_world")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[string]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != "hello_world" {
			t.Errorf("expected hello_world, got %v", result)
		}
	})

	t.Run("returns int when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "12345")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[int]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 12345 {
			t.Errorf("expected 12345, got %v", result)
		}
	})

	t.Run("returns uint when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "12345")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[uint]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 12345 {
			t.Errorf("expected 12345, got %v", result)
		}
	})

	t.Run("returns uint8 when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "255")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[uint8]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 255 {
			t.Errorf("expected 255, got %v", result)
		}
	})

	t.Run("returns uint16 when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "65535")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[uint16]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 65535 {
			t.Errorf("expected 65535, got %v", result)
		}
	})

	t.Run("returns uint32 when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "4294967295")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[uint32]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 4294967295 {
			t.Errorf("expected 4294967295, got %v", result)
		}
	})

	t.Run("returns uint64 when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "18446744073709551615")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[uint64]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 18446744073709551615 {
			t.Errorf("expected 18446744073709551615, got %v", result)
		}
	})

	t.Run("returns error when env var is not set", func(t *testing.T) {
		os.Unsetenv("TEST_VAR_NOT_SET")

		result, err := TryGetEnvironmentVariable[int]("TEST_VAR_NOT_SET")
		if err == nil {
			t.Error("expected error when env var is not set")
		}

		if result != 0 {
			t.Errorf("expected zero value, got %v", result)
		}
	})

	t.Run("returns error when env var is empty", func(t *testing.T) {
		os.Setenv("TEST_VAR_EMPTY", "")
		defer os.Unsetenv("TEST_VAR_EMPTY")

		result, err := TryGetEnvironmentVariable[int]("TEST_VAR_EMPTY")
		if err == nil {
			t.Error("expected error when env var is empty")
		}

		if result != 0 {
			t.Errorf("expected zero value, got %v", result)
		}
	})

	t.Run("returns error when bool value is invalid", func(t *testing.T) {
		os.Setenv("TEST_VAR", "not_a_bool")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[bool]("TEST_VAR")
		if err == nil {
			t.Error("expected error for invalid bool")
		}

		if result != false {
			t.Errorf("expected false (zero value), got %v", result)
		}
	})

	t.Run("returns error when int value is invalid", func(t *testing.T) {
		os.Setenv("TEST_VAR", "not_a_number")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[int]("TEST_VAR")
		if err == nil {
			t.Error("expected error for invalid int")
		}

		if result != 0 {
			t.Errorf("expected 0 (zero value), got %v", result)
		}
	})

	t.Run("returns error when uint value is negative", func(t *testing.T) {
		os.Setenv("TEST_VAR", "-1")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[uint]("TEST_VAR")
		if err == nil {
			t.Error("expected error for negative uint")
		}

		if result != 0 {
			t.Errorf("expected 0 (zero value), got %v", result)
		}
	})

	t.Run("returns error when int8 overflows", func(t *testing.T) {
		os.Setenv("TEST_VAR", "128")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[int8]("TEST_VAR")
		if err == nil {
			t.Error("expected error for int8 overflow")
		}

		if result != 0 {
			t.Errorf("expected 0 (zero value), got %v", result)
		}
	})

	t.Run("returns error when uint8 overflows", func(t *testing.T) {
		os.Setenv("TEST_VAR", "256")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[uint8]("TEST_VAR")
		if err == nil {
			t.Error("expected error for uint8 overflow")
		}

		if result != 0 {
			t.Errorf("expected 0 (zero value), got %v", result)
		}
	})

	t.Run("returns error when float value is invalid", func(t *testing.T) {
		os.Setenv("TEST_VAR", "not_a_float")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[float64]("TEST_VAR")
		if err == nil {
			t.Error("expected error for invalid float")
		}

		if result != 0.0 {
			t.Errorf("expected 0.0 (zero value), got %v", result)
		}
	})

	t.Run("handles negative int values", func(t *testing.T) {
		os.Setenv("TEST_VAR", "-127")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[int8]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != -127 {
			t.Errorf("expected -127, got %v", result)
		}
	})

	t.Run("handles negative float values", func(t *testing.T) {
		os.Setenv("TEST_VAR", "-3.14159")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[float64]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != -3.14159 {
			t.Errorf("expected -3.14159, got %v", result)
		}
	})

	t.Run("handles zero values", func(t *testing.T) {
		os.Setenv("TEST_VAR", "0")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[int]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 0 {
			t.Errorf("expected 0, got %v", result)
		}
	})

	t.Run("handles bool 1 and 0", func(t *testing.T) {
		os.Setenv("TEST_VAR", "1")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariable[bool]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if !result {
			t.Error("expected true for '1'")
		}

		os.Setenv("TEST_VAR", "0")
		result, err = TryGetEnvironmentVariable[bool]("TEST_VAR")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result {
			t.Error("expected false for '0'")
		}
	})
}

func TestTryGetEnvironmentVariableOrDefault(t *testing.T) {
	t.Run("returns value when string env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "custom_value")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR", "default")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != "custom_value" {
			t.Errorf("expected custom_value, got %v", result)
		}
	})

	t.Run("returns default when string env var is not set", func(t *testing.T) {
		os.Unsetenv("TEST_VAR_NOT_SET")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR_NOT_SET", "default_value")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != "default_value" {
			t.Errorf("expected default_value, got %v", result)
		}
	})

	t.Run("returns default when env var is empty", func(t *testing.T) {
		os.Setenv("TEST_VAR_EMPTY", "")
		defer os.Unsetenv("TEST_VAR_EMPTY")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR_EMPTY", "default_value")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != "default_value" {
			t.Errorf("expected default_value, got %v", result)
		}
	})

	t.Run("returns int value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "42")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR", 10)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 42 {
			t.Errorf("expected 42, got %v", result)
		}
	})

	t.Run("returns default int when env var is not set", func(t *testing.T) {
		os.Unsetenv("TEST_VAR_NOT_SET")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR_NOT_SET", 99)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 99 {
			t.Errorf("expected 99, got %v", result)
		}
	})

	t.Run("returns bool true when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "true")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR", false)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if !result {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns bool false when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "false")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR", true)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result {
			t.Error("expected false, got true")
		}
	})

	t.Run("returns default bool when env var is not set", func(t *testing.T) {
		os.Unsetenv("TEST_VAR_NOT_SET")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR_NOT_SET", true)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if !result {
			t.Error("expected true (default), got false")
		}
	})

	t.Run("returns float64 value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "3.14159")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR", 1.0)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 3.14159 {
			t.Errorf("expected 3.14159, got %v", result)
		}
	})

	t.Run("returns default float64 when env var is not set", func(t *testing.T) {
		os.Unsetenv("TEST_VAR_NOT_SET")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR_NOT_SET", 2.71828)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 2.71828 {
			t.Errorf("expected 2.71828, got %v", result)
		}
	})

	t.Run("returns int8 value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "100")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault[int8]("TEST_VAR", 50)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 100 {
			t.Errorf("expected 100, got %v", result)
		}
	})

	t.Run("returns int16 value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "1000")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault[int16]("TEST_VAR", 500)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 1000 {
			t.Errorf("expected 1000, got %v", result)
		}
	})

	t.Run("returns int32 value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "100000")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault[int32]("TEST_VAR", 50000)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 100000 {
			t.Errorf("expected 100000, got %v", result)
		}
	})

	t.Run("returns int64 value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "9223372036854775807")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault[int64]("TEST_VAR", 0)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 9223372036854775807 {
			t.Errorf("expected max int64, got %v", result)
		}
	})

	t.Run("returns uint value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "12345")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault[uint]("TEST_VAR", 0)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 12345 {
			t.Errorf("expected 12345, got %v", result)
		}
	})

	t.Run("returns uint8 value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "200")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault[uint8]("TEST_VAR", 0)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 200 {
			t.Errorf("expected 200, got %v", result)
		}
	})

	t.Run("returns uint16 value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "50000")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault[uint16]("TEST_VAR", 0)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 50000 {
			t.Errorf("expected 50000, got %v", result)
		}
	})

	t.Run("returns uint32 value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "3000000000")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault[uint32]("TEST_VAR", 0)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 3000000000 {
			t.Errorf("expected 3000000000, got %v", result)
		}
	})

	t.Run("returns uint64 value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "18446744073709551615")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault[uint64]("TEST_VAR", 0)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 18446744073709551615 {
			t.Errorf("expected max uint64, got %v", result)
		}
	})

	t.Run("returns float32 value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "3.14")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault[float32]("TEST_VAR", 0.0)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result < 3.13 || result > 3.15 {
			t.Errorf("expected ~3.14, got %v", result)
		}
	})

	t.Run("returns error when int value is invalid", func(t *testing.T) {
		os.Setenv("TEST_VAR", "not_a_number")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR", 10)
		if err == nil {
			t.Error("expected error for invalid int")
		}

		if result != 0 {
			t.Errorf("expected 0 (zero value), got %v", result)
		}
	})

	t.Run("returns error when bool value is invalid", func(t *testing.T) {
		os.Setenv("TEST_VAR", "not_a_bool")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR", true)
		if err == nil {
			t.Error("expected error for invalid bool")
		}

		if result != false {
			t.Errorf("expected false (zero value), got %v", result)
		}
	})

	t.Run("returns error when float value is invalid", func(t *testing.T) {
		os.Setenv("TEST_VAR", "not_a_float")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR", 1.5)
		if err == nil {
			t.Error("expected error for invalid float")
		}

		if result != 0.0 {
			t.Errorf("expected 0.0 (zero value), got %v", result)
		}
	})

	t.Run("returns error when int8 overflows", func(t *testing.T) {
		os.Setenv("TEST_VAR", "128")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault[int8]("TEST_VAR", 10)
		if err == nil {
			t.Error("expected error for int8 overflow")
		}

		if result != 0 {
			t.Errorf("expected 0 (zero value), got %v", result)
		}
	})

	t.Run("returns error when uint value is negative", func(t *testing.T) {
		os.Setenv("TEST_VAR", "-1")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault[uint]("TEST_VAR", 10)
		if err == nil {
			t.Error("expected error for negative uint")
		}

		if result != 0 {
			t.Errorf("expected 0 (zero value), got %v", result)
		}
	})

	t.Run("handles zero env var value returning actual zero not default", func(t *testing.T) {
		os.Setenv("TEST_VAR", "0")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR", 999)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != 0 {
			t.Errorf("expected 0 (from env), got %v", result)
		}
	})

	t.Run("handles false env var value returning actual false not default", func(t *testing.T) {
		os.Setenv("TEST_VAR", "false")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR", true)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != false {
			t.Errorf("expected false (from env), got %v", result)
		}
	})

	t.Run("handles negative values", func(t *testing.T) {
		os.Setenv("TEST_VAR", "-42")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR", 10)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != -42 {
			t.Errorf("expected -42, got %v", result)
		}
	})

	t.Run("handles whitespace in string values", func(t *testing.T) {
		os.Setenv("TEST_VAR", "  hello  ")
		defer os.Unsetenv("TEST_VAR")

		result, err := TryGetEnvironmentVariableOrDefault("TEST_VAR", "default")
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if result != "  hello  " {
			t.Errorf("expected '  hello  ', got '%v'", result)
		}
	})
}
