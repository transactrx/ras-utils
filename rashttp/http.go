// Package rashttp provides HTTP helper functions for request parsing, response writing,
// and common web application patterns.
//
// It includes utilities for extracting client information from requests (IP, bearer tokens),
// detecting request types (AJAX, HTMX), and writing JSON responses with standard status codes.
package rashttp

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// DefaultMaxBodySize is the default maximum request body size for [DecodeJSON] (1MB).
const DefaultMaxBodySize = 1 << 20

// GetFullRequestURL reconstructs the full URL from a request, including scheme and host.
// It respects X-Forwarded-Proto and X-Forwarded-Host headers for proxied requests.
func GetFullRequestURL(r *http.Request) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
		if r.TLS != nil {
			scheme = "https"
		}
	}
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	return scheme + "://" + host + r.URL.RequestURI()
}

// GetClientIP extracts the client IP address from the request,
// checking X-Forwarded-For and X-Real-IP headers first (for proxied requests).
func GetClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// GetBearerToken extracts the bearer token from the Authorization header.
// Returns empty string if not present or not a bearer token.
func GetBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(auth[len(prefix):])
}

// IsHTMX reports whether the request was made by HTMX (checks HX-Request header).
func IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// IsAjax reports whether the request appears to be an AJAX/XHR request (checks X-Requested-With header).
func IsAjax(r *http.Request) bool {
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// WriteJSON encodes data as JSON and writes it to the response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// ErrorResponse is the standard JSON structure for error responses.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// WriteError writes a JSON [ErrorResponse] with the given status code and message.
func WriteError(w http.ResponseWriter, status int, message string) error {
	return WriteJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

// OK writes a 200 response with JSON data.
func OK(w http.ResponseWriter, data any) error {
	return WriteJSON(w, http.StatusOK, data)
}

// Created writes a 201 response with JSON data.
func Created(w http.ResponseWriter, data any) error {
	return WriteJSON(w, http.StatusCreated, data)
}

// Accepted writes a 202 response with JSON data.
func Accepted(w http.ResponseWriter, data any) error {
	return WriteJSON(w, http.StatusAccepted, data)
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest writes a 400 error response.
func BadRequest(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusBadRequest, message)
}

// Unauthorized writes a 401 error response.
func Unauthorized(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusUnauthorized, message)
}

// Forbidden writes a 403 error response.
func Forbidden(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusForbidden, message)
}

// NotFound writes a 404 error response.
func NotFound(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusNotFound, message)
}

// Conflict writes a 409 error response.
func Conflict(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusConflict, message)
}

// UnprocessableEntity writes a 422 error response.
func UnprocessableEntity(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusUnprocessableEntity, message)
}

// TooManyRequests writes a 429 error response.
func TooManyRequests(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusTooManyRequests, message)
}

// InternalServerError writes a 500 error response.
func InternalServerError(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusInternalServerError, message)
}

// ServiceUnavailable writes a 503 error response.
func ServiceUnavailable(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusServiceUnavailable, message)
}

// DecodeJSON decodes JSON from the request body into a value of type T.
// It limits body size to maxSize; use [DefaultMaxBodySize] if unsure.
func DecodeJSON[T any](r *http.Request, maxSize int64) (T, error) {
	var v T
	body := http.MaxBytesReader(nil, r.Body, maxSize)
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		return v, err
	}
	err = json.Unmarshal(data, &v)
	return v, err
}

// QueryString returns the query parameter value or the default if not present.
func QueryString(r *http.Request, key, defaultValue string) string {
	if v := r.URL.Query().Get(key); v != "" {
		return v
	}
	return defaultValue
}

// QueryInt returns the query parameter as an int or the default if not present or invalid.
func QueryInt(r *http.Request, key string, defaultValue int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return i
}

// HealthCheckFunc is a function that returns nil if healthy, or an error otherwise.
type HealthCheckFunc func() error

// HealthHandler returns an [http.HandlerFunc] that runs the health check.
// It returns 200 OK if healthy, or 503 Service Unavailable if the check returns an error.
func HealthHandler(check HealthCheckFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := check(); err != nil {
			WriteError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
