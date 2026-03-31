package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

func GetEnvironmentVariableOrDefault(key string, defaultValue string) string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	return valueStr
}

func GetEnvironmentVariableOrPanic(key string, panicMessage string) string {
	value := os.Getenv(key)
	if value == "" {
		slog.Error("environment variable key is missing", "key missing", key)
		panic(panicMessage)
	}
	return value
}

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
