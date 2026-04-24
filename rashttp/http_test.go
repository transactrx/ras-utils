package rashttp

import (
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetFullRequestURL(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*http.Request)
		expected string
	}{
		{
			name: "basic request",
			setup: func(r *http.Request) {
				r.Host = "example.com"
				r.URL.Path = "/api/test"
			},
			expected: "http://example.com/api/test",
		},
		{
			name: "with query string",
			setup: func(r *http.Request) {
				r.Host = "example.com"
				r.URL.Path = "/api/test"
				r.URL.RawQuery = "foo=bar"
			},
			expected: "http://example.com/api/test?foo=bar",
		},
		{
			name: "with X-Forwarded headers",
			setup: func(r *http.Request) {
				r.Host = "internal.local"
				r.URL.Path = "/api/test"
				r.Header.Set("X-Forwarded-Proto", "https")
				r.Header.Set("X-Forwarded-Host", "public.example.com")
			},
			expected: "https://public.example.com/api/test",
		},
		{
			name: "TLS request",
			setup: func(r *http.Request) {
				r.Host = "example.com"
				r.URL.Path = "/secure"
				r.TLS = &tls.ConnectionState{}
			},
			expected: "https://example.com/secure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header = make(http.Header)
			tt.setup(r)
			got := GetFullRequestURL(r)
			if got != tt.expected {
				t.Errorf("GetFullRequestURL() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*http.Request)
		expected string
	}{
		{
			name: "X-Forwarded-For single IP",
			setup: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "203.0.113.50")
			},
			expected: "203.0.113.50",
		},
		{
			name: "X-Forwarded-For multiple IPs",
			setup: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18, 150.172.238.178")
			},
			expected: "203.0.113.50",
		},
		{
			name: "X-Real-IP",
			setup: func(r *http.Request) {
				r.Header.Set("X-Real-IP", "203.0.113.100")
			},
			expected: "203.0.113.100",
		},
		{
			name: "X-Forwarded-For takes precedence over X-Real-IP",
			setup: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "203.0.113.50")
				r.Header.Set("X-Real-IP", "203.0.113.100")
			},
			expected: "203.0.113.50",
		},
		{
			name: "fallback to RemoteAddr",
			setup: func(r *http.Request) {
				r.RemoteAddr = "192.168.1.1:12345"
			},
			expected: "192.168.1.1",
		},
		{
			name: "RemoteAddr without port",
			setup: func(r *http.Request) {
				r.RemoteAddr = "192.168.1.1"
			},
			expected: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header = make(http.Header)
			r.RemoteAddr = ""
			tt.setup(r)
			got := GetClientIP(r)
			if got != tt.expected {
				t.Errorf("GetClientIP() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name     string
		auth     string
		expected string
	}{
		{"valid bearer token", "Bearer abc123", "abc123"},
		{"bearer lowercase", "bearer abc123", "abc123"},
		{"bearer mixed case", "BEARER abc123", "abc123"},
		{"no authorization header", "", ""},
		{"basic auth", "Basic dXNlcjpwYXNz", ""},
		{"bearer with extra spaces", "Bearer   token123  ", "token123"},
		{"just bearer prefix", "Bearer ", ""},
		{"bearer no space", "Bearerabc123", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.auth != "" {
				r.Header.Set("Authorization", tt.auth)
			}
			got := GetBearerToken(r)
			if got != tt.expected {
				t.Errorf("GetBearerToken() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIsHTMX(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected bool
	}{
		{"htmx request", "true", true},
		{"not htmx", "false", false},
		{"no header", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				r.Header.Set("HX-Request", tt.header)
			}
			got := IsHTMX(r)
			if got != tt.expected {
				t.Errorf("IsHTMX() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsAjax(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected bool
	}{
		{"ajax request", "XMLHttpRequest", true},
		{"not ajax", "SomethingElse", false},
		{"no header", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				r.Header.Set("X-Requested-With", tt.header)
			}
			got := IsAjax(r)
			if got != tt.expected {
				t.Errorf("IsAjax() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "hello"}

	err := WriteJSON(w, http.StatusOK, data)
	if err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
	expected := `{"message":"hello"}`
	got := strings.TrimSpace(w.Body.String())
	if got != expected {
		t.Errorf("body = %q, want %q", got, expected)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	err := WriteError(w, http.StatusBadRequest, "invalid input")
	if err != nil {
		t.Fatalf("WriteError() error = %v", err)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	expected := `{"error":"Bad Request","message":"invalid input"}`
	got := strings.TrimSpace(w.Body.String())
	if got != expected {
		t.Errorf("body = %q, want %q", got, expected)
	}
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	NoContent(w)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if w.Body.Len() != 0 {
		t.Errorf("body should be empty, got %q", w.Body.String())
	}
}

func TestDecodeJSON(t *testing.T) {
	type payload struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	t.Run("valid JSON", func(t *testing.T) {
		body := strings.NewReader(`{"name":"test","value":42}`)
		r := httptest.NewRequest(http.MethodPost, "/", body)

		got, err := DecodeJSON[payload](r, DefaultMaxBodySize)
		if err != nil {
			t.Fatalf("DecodeJSON() error = %v", err)
		}
		if got.Name != "test" || got.Value != 42 {
			t.Errorf("DecodeJSON() = %+v, want {Name:test Value:42}", got)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		body := strings.NewReader(`{invalid}`)
		r := httptest.NewRequest(http.MethodPost, "/", body)

		_, err := DecodeJSON[payload](r, DefaultMaxBodySize)
		if err == nil {
			t.Error("DecodeJSON() expected error for invalid JSON")
		}
	})

	t.Run("body too large", func(t *testing.T) {
		body := strings.NewReader(`{"name":"test","value":42}`)
		r := httptest.NewRequest(http.MethodPost, "/", body)

		_, err := DecodeJSON[payload](r, 5)
		if err == nil {
			t.Error("DecodeJSON() expected error for oversized body")
		}
	})
}

func TestQueryString(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		key          string
		defaultValue string
		expected     string
	}{
		{"present", "?foo=bar", "foo", "default", "bar"},
		{"missing", "?other=value", "foo", "default", "default"},
		{"empty value", "?foo=", "foo", "default", "default"},
		{"no query", "", "foo", "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)
			got := QueryString(r, tt.key, tt.defaultValue)
			if got != tt.expected {
				t.Errorf("QueryString() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestQueryInt(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		key          string
		defaultValue int
		expected     int
	}{
		{"valid int", "?page=5", "page", 1, 5},
		{"missing", "?other=5", "page", 1, 1},
		{"invalid int", "?page=abc", "page", 1, 1},
		{"negative int", "?page=-10", "page", 1, -10},
		{"zero", "?page=0", "page", 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)
			got := QueryInt(r, tt.key, tt.defaultValue)
			if got != tt.expected {
				t.Errorf("QueryInt() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestStatusHelpers(t *testing.T) {
	tests := []struct {
		name           string
		fn             func(http.ResponseWriter) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "OK",
			fn:             func(w http.ResponseWriter) error { return OK(w, map[string]string{"id": "1"}) },
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"1"}`,
		},
		{
			name:           "Created",
			fn:             func(w http.ResponseWriter) error { return Created(w, map[string]string{"id": "2"}) },
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":"2"}`,
		},
		{
			name:           "Accepted",
			fn:             func(w http.ResponseWriter) error { return Accepted(w, map[string]string{"job": "abc"}) },
			expectedStatus: http.StatusAccepted,
			expectedBody:   `{"job":"abc"}`,
		},
		{
			name:           "BadRequest",
			fn:             func(w http.ResponseWriter) error { return BadRequest(w, "bad input") },
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Bad Request","message":"bad input"}`,
		},
		{
			name:           "Unauthorized",
			fn:             func(w http.ResponseWriter) error { return Unauthorized(w, "no token") },
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Unauthorized","message":"no token"}`,
		},
		{
			name:           "Forbidden",
			fn:             func(w http.ResponseWriter) error { return Forbidden(w, "not allowed") },
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"Forbidden","message":"not allowed"}`,
		},
		{
			name:           "NotFound",
			fn:             func(w http.ResponseWriter) error { return NotFound(w, "missing") },
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Not Found","message":"missing"}`,
		},
		{
			name:           "Conflict",
			fn:             func(w http.ResponseWriter) error { return Conflict(w, "duplicate") },
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"Conflict","message":"duplicate"}`,
		},
		{
			name:           "UnprocessableEntity",
			fn:             func(w http.ResponseWriter) error { return UnprocessableEntity(w, "validation failed") },
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"error":"Unprocessable Entity","message":"validation failed"}`,
		},
		{
			name:           "TooManyRequests",
			fn:             func(w http.ResponseWriter) error { return TooManyRequests(w, "slow down") },
			expectedStatus: http.StatusTooManyRequests,
			expectedBody:   `{"error":"Too Many Requests","message":"slow down"}`,
		},
		{
			name:           "InternalServerError",
			fn:             func(w http.ResponseWriter) error { return InternalServerError(w, "something broke") },
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal Server Error","message":"something broke"}`,
		},
		{
			name:           "ServiceUnavailable",
			fn:             func(w http.ResponseWriter) error { return ServiceUnavailable(w, "down") },
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `{"error":"Service Unavailable","message":"down"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := tt.fn(w)
			if err != nil {
				t.Fatalf("%s() error = %v", tt.name, err)
			}
			if w.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectedStatus)
			}
			got := strings.TrimSpace(w.Body.String())
			if got != tt.expectedBody {
				t.Errorf("body = %q, want %q", got, tt.expectedBody)
			}
		})
	}
}

func TestHealthHandler(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		check := func() error { return nil }
		handler := HealthHandler(check)

		r := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		handler(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
		body, _ := io.ReadAll(w.Body)
		if !strings.Contains(string(body), `"status":"ok"`) {
			t.Errorf("body = %q, want status:ok", string(body))
		}
	})

	t.Run("unhealthy", func(t *testing.T) {
		check := func() error { return errors.New("db connection failed") }
		handler := HealthHandler(check)

		r := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		handler(w, r)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
		}
		body, _ := io.ReadAll(w.Body)
		if !strings.Contains(string(body), "db connection failed") {
			t.Errorf("body = %q, want error message", string(body))
		}
	})
}
