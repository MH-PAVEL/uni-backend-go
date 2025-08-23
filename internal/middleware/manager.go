package middleware

import "net/http"

// Chain applies multiple middlewares in order
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// loop backwards so middleware order is preserved
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
