package rasconfig

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

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
