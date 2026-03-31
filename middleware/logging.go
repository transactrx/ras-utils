package middleware

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

func SetUpLogger() {
	setDefaultLoggerLevel()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}).WithAttrs([]slog.Attr{slog.String("version", runtime.Version())}))

	slog.SetDefault(logger)
}

func setDefaultLoggerLevel() {
	defaultMinLevelString := os.Getenv("LOG_LEVEL")

	defaultMinLevel, err := parseLogLevel(defaultMinLevelString)
	if err != nil {
		defaultMinLevel = slog.LevelInfo
		slog.Error("error parsing LOG_LEVEL env var, using default level info", "error", err)
	}

	level.Set(defaultMinLevel)
}

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

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func LoggingMiddleware(l *slog.Logger) func(next http.Handler) http.Handler {
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

			if path != "/isAvailable" {
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
