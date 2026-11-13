package rasconfig

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type primitive interface {
	~bool | ~string |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~complex64 | ~complex128
}

// ErrEnvNotFound is returned when an environment variable is not set or empty.
var ErrEnvNotFound = errors.New("environment variable not found")

// GetEnvironmentVariableOrDefault returns the value of the environment variable named by key,
// or defaultValue if the variable is not set or empty.
func GetEnvironmentVariableOrDefault(key string, defaultValue string) string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	return valueStr
}

// GetEnvironmentVariableOrPanic returns the value of the environment variable named by key.
// It panics with panicMessage if the variable is not set or empty.
func GetEnvironmentVariableOrPanic(key string, panicMessage string) string {
	value := os.Getenv(key)
	if value == "" {
		slog.Error("environment variable key is missing", "key missing", key)
		panic(panicMessage)
	}
	return value
}

// GetEnvironmentVariableOrDefaultInt returns the integer value of the environment variable named by key,
// or defaultValue if the variable is not set, empty, or not a valid integer.
func GetEnvironmentVariableOrDefaultInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	valueInt, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return valueInt
}

// GetEnvironmentVariableOrDefaultDuration returns the duration value of the environment variable named by key,
// or the parsed defaultValue if the variable is not set, empty, or not a valid duration string.
func GetEnvironmentVariableOrDefaultDuration(key string, defaultValue string) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		duration, _ := time.ParseDuration(defaultValue)
		return duration
	}

	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		slog.Warn("Invalid value for environment variable.", "key", key, "defaultValue", defaultValue)
		duration, _ = time.ParseDuration(defaultValue)
	}

	return duration
}

// TryGetEnvironmentVariableOrDefault returns the parsed value of the environment variable named by key,
// or defaultValue if the variable is not set or empty.
// Returns an error if the variable is set but cannot be parsed as type T.
func TryGetEnvironmentVariableOrDefault[T primitive](key string, defaultValue T) (T, error) {
	val, err := TryGetEnvironmentVariable[T](key)
	if errors.Is(err, ErrEnvNotFound) {
		return defaultValue, nil
	}
	return val, err
}

// TryGetEnvironmentVariable returns the parsed value of the environment variable named by key.
// Returns ErrEnvNotFound if the variable is not set or empty.
// Returns a parse error if the value cannot be converted to type T.
func TryGetEnvironmentVariable[T primitive](key string) (T, error) {
	s := os.Getenv(key)
	var zero T
	if s == "" {
		return zero, ErrEnvNotFound
	}
	result, err := parseVariable[T](s)
	if err != nil {
		return zero, fmt.Errorf("unable to parse environment variable %q: %w", key, err)
	}
	return result, nil
}

func parseVariable[T primitive](s string) (T, error) {
	var result T

	// Switch on the pointer to result cast as 'any' to mutate it directly
	switch ptr := any(&result).(type) {
	case *string:
		*ptr = s
	case *bool:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return result, err
		}
		*ptr = v
	case *int:
		v, err := strconv.ParseInt(s, 10, 0)
		if err != nil {
			return result, err
		}
		*ptr = int(v)
	case *int8:
		v, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			return result, err
		}
		*ptr = int8(v)
	case *int16:
		v, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return result, err
		}
		*ptr = int16(v)
	case *int32:
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return result, err
		}
		*ptr = int32(v)
	case *int64:
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return result, err
		}
		*ptr = v
	case *uint:
		v, err := strconv.ParseUint(s, 10, 0)
		if err != nil {
			return result, err
		}
		*ptr = uint(v)
	case *uint8:
		v, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			return result, err
		}
		*ptr = uint8(v)
	case *uint16:
		v, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return result, err
		}
		*ptr = uint16(v)
	case *uint32:
		v, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return result, err
		}
		*ptr = uint32(v)
	case *uint64:
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return result, err
		}
		*ptr = v
	case *float32:
		v, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return result, err
		}
		*ptr = float32(v)
	case *float64:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return result, err
		}
		*ptr = v
	default:
		// Fallback using fmt.Sscan for custom underlying types (e.g., type MyInt int)
		if _, err := fmt.Sscan(s, &result); err != nil {
			return result, fmt.Errorf("unsupported type or invalid value %q: %w", s, err)
		}
	}

	return result, nil
}
