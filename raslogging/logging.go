// Package raslogging provides HTTP request logging middleware with panic recovery
// and structured JSON logging via [log/slog].
package raslogging

import (
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

var (
	level = new(slog.LevelVar)
)

// SetUpLogger initializes the default [slog] logger with JSON formatting.
// It reads the LOG_LEVEL environment variable to set the minimum log level.
func SetUpLogger() {
	setDefaultLoggerLevel()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}).WithAttrs([]slog.Attr{slog.String("version", runtime.Version())}))

	slog.SetDefault(logger)
}

// setDefaultLoggerLevel sets the log level from the LOG_LEVEL environment variable.
func setDefaultLoggerLevel() {
	defaultMinLevelString := os.Getenv("LOG_LEVEL")

	defaultMinLevel, err := parseLogLevel(defaultMinLevelString)
	if err != nil {
		defaultMinLevel = slog.LevelInfo
		slog.Error("error parsing LOG_LEVEL env var, using default level info", "error", err)
	}

	level.Set(defaultMinLevel)
}

// parseLogLevel parses a log level string (e.g., "debug", "info", "warn", "error").
func parseLogLevel(s string) (slog.Level, error) {
	var level slog.Level
	var err = level.UnmarshalText([]byte(s))
	return level, err
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// wrapResponseWriter wraps an [http.ResponseWriter] to capture the status code.
func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

// Status returns the HTTP status code written to the response.
func (rw *responseWriter) Status() int {
	return rw.status
}

// WriteHeader implements [http.ResponseWriter] and captures the status code.
func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

// LoggingMiddleware returns HTTP middleware that logs requests and recovers from panics.
// Paths in skipPaths are not logged (useful for health checks).
func LoggingMiddleware(l *slog.Logger, skipPaths ...string) func(next http.Handler) http.Handler {
	skipSet := make(map[string]struct{}, len(skipPaths))
	for _, p := range skipPaths {
		skipSet[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					l.Error("error in http request middleware",
						"err", err,
						"trace", debug.Stack(),
					)
				}
			}()

			start := time.Now()
			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)
			path := r.URL.EscapedPath()

			if _, skip := skipSet[path]; !skip {
				l.Debug("http request",
					"status", wrapped.status,
					"method", r.Method,
					"path", path,
					"duration in ms", time.Since(start).Milliseconds(),
				)
			}
		}

		return http.HandlerFunc(fn)
	}
}
