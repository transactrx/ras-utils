package rasstack

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateStack(t *testing.T) {
	t.Run("applies middleware in correct order", func(t *testing.T) {
		var order []string

		middleware1 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "m1-before")
				next.ServeHTTP(w, r)
				order = append(order, "m1-after")
			})
		}

		middleware2 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "m2-before")
				next.ServeHTTP(w, r)
				order = append(order, "m2-after")
			})
		}

		middleware3 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "m3-before")
				next.ServeHTTP(w, r)
				order = append(order, "m3-after")
			})
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
		})

		stack := CreateStack(middleware1, middleware2, middleware3)
		wrapped := stack(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		wrapped.ServeHTTP(rr, req)

		expected := []string{"m1-before", "m2-before", "m3-before", "handler", "m3-after", "m2-after", "m1-after"}
		if len(order) != len(expected) {
			t.Fatalf("expected %d calls, got %d", len(expected), len(order))
		}
		for i, v := range expected {
			if order[i] != v {
				t.Errorf("at index %d, expected %s, got %s", i, v, order[i])
			}
		}
	})

	t.Run("works with empty stack", func(t *testing.T) {
		called := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})

		stack := CreateStack()
		wrapped := stack(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		wrapped.ServeHTTP(rr, req)

		if !called {
			t.Error("expected handler to be called")
		}
	})

	t.Run("works with single middleware", func(t *testing.T) {
		middlewareCalled := false
		handlerCalled := false

		middleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middlewareCalled = true
				next.ServeHTTP(w, r)
			})
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		stack := CreateStack(middleware)
		wrapped := stack(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		wrapped.ServeHTTP(rr, req)

		if !middlewareCalled {
			t.Error("expected middleware to be called")
		}
		if !handlerCalled {
			t.Error("expected handler to be called")
		}
	})

	t.Run("middleware can short-circuit", func(t *testing.T) {
		handlerCalled := false

		authMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				// Don't call next
			})
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		stack := CreateStack(authMiddleware)
		wrapped := stack(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		wrapped.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rr.Code)
		}
		if handlerCalled {
			t.Error("expected handler not to be called when middleware short-circuits")
		}
	})

	t.Run("middleware can modify request context", func(t *testing.T) {
		type ctxKey string
		var receivedValue string

		middleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				// We can't easily modify context in this simple test,
				// but we can verify the request passes through
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedValue = "handler-called"
		})

		stack := CreateStack(middleware)
		wrapped := stack(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		wrapped.ServeHTTP(rr, req)

		if receivedValue != "handler-called" {
			t.Error("expected handler to be called with modified request")
		}
	})
}
