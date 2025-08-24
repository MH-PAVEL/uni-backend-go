package routes

import (
	"net/http"

	"github.com/MH-PAVEL/uni-backend-go/internal/handlers"
	"github.com/MH-PAVEL/uni-backend-go/internal/middleware"
)

func RegisterAuthRoutes(mux *http.ServeMux) {
	// Public

	mux.Handle("POST /api/v1/auth/signup", middleware.Chain(http.HandlerFunc(handlers.Signup)))
	mux.Handle("POST /api/v1/auth/login",  middleware.Chain(http.HandlerFunc(handlers.Login)))
	mux.Handle("POST /api/v1/auth/logout", middleware.Chain(http.HandlerFunc(handlers.Logout)))
	mux.Handle("POST /api/v1/auth/refresh", middleware.Chain(http.HandlerFunc(handlers.Refresh)))



	// Protected
	mux.Handle("GET /api/v1/auth/me",
		middleware.Chain(
			http.HandlerFunc(handlers.GetMe),
			middleware.AuthMiddleware,
		),
	)
}
