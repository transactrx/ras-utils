package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestSetUpLogger(t *testing.T) {
	t.Run("sets up logger without panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SetUpLogger panicked: %v", r)
			}
		}()

		SetUpLogger()
	})

	t.Run("respects LOG_LEVEL env var", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "DEBUG")
		defer os.Unsetenv("LOG_LEVEL")

		SetUpLogger()
	})

	t.Run("handles invalid LOG_LEVEL", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "INVALID_LEVEL")
		defer os.Unsetenv("LOG_LEVEL")

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SetUpLogger panicked on invalid level: %v", r)
			}
		}()

		SetUpLogger()
	})
}

func TestParseLogLevel(t *testing.T) {
	testCases := []struct {
		input       string
		expected    slog.Level
		shouldError bool
	}{
		{"DEBUG", slog.LevelDebug, false},
		{"INFO", slog.LevelInfo, false},
		{"WARN", slog.LevelWarn, false},
		{"ERROR", slog.LevelError, false},
		{"debug", slog.LevelDebug, false},
		{"info", slog.LevelInfo, false},
		{"INVALID", slog.Level(0), true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			level, err := parseLogLevel(tc.input)
			if tc.shouldError {
				if err == nil {
					t.Errorf("expected error for input %s", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %s: %v", tc.input, err)
				}
				if level != tc.expected {
					t.Errorf("for input %s, expected %v, got %v", tc.input, tc.expected, level)
				}
			}
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	t.Run("logs requests", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := LoggingMiddleware(logger)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
		rr := httptest.NewRecorder()

		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}

		logOutput := buf.String()
		if logOutput == "" {
			t.Error("expected log output")
		}
	})

	t.Run("skips specified paths", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := LoggingMiddleware(logger, "/health", "/ready")
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()

		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}

		logOutput := buf.String()
		if logOutput != "" {
			t.Errorf("expected no log output for skipped path, got: %s", logOutput)
		}
	})

	t.Run("logs non-skipped paths", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := LoggingMiddleware(logger, "/health")
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		rr := httptest.NewRecorder()

		wrapped.ServeHTTP(rr, req)

		logOutput := buf.String()
		if logOutput == "" {
			t.Error("expected log output for non-skipped path")
		}
	})

	t.Run("recovers from panic", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		middleware := LoggingMiddleware(logger)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/panic-path", nil)
		rr := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic was not recovered: %v", r)
			}
		}()

		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500 after panic, got %d", rr.Code)
		}
	})

	t.Run("captures response status", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		middleware := LoggingMiddleware(logger)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
		rr := httptest.NewRecorder()

		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rr.Code)
		}
	})
}

func TestResponseWriter(t *testing.T) {
	t.Run("captures status code", func(t *testing.T) {
		rr := httptest.NewRecorder()
		wrapped := wrapResponseWriter(rr)

		wrapped.WriteHeader(http.StatusCreated)

		if wrapped.Status() != http.StatusCreated {
			t.Errorf("expected status 201, got %d", wrapped.Status())
		}
	})

	t.Run("only writes header once", func(t *testing.T) {
		rr := httptest.NewRecorder()
		wrapped := wrapResponseWriter(rr)

		wrapped.WriteHeader(http.StatusCreated)
		wrapped.WriteHeader(http.StatusOK)

		if wrapped.Status() != http.StatusCreated {
			t.Errorf("expected first status 201 to be preserved, got %d", wrapped.Status())
		}
	})
}
