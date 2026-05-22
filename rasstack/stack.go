// Package rasstack provides middleware composition utilities for HTTP handlers.
package rasstack

import "net/http"

// Middleware is a function that wraps an [http.Handler].
type Middleware func(http.Handler) http.Handler

// CreateStack composes multiple middleware into a single [Middleware].
// Middleware are applied in the order provided, so the first middleware in the list
// is the outermost (first to receive the request, last to see the response).
func CreateStack(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			x := xs[i]
			next = x(next)
		}
		return next
	}
}
