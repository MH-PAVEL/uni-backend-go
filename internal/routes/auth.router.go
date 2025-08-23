package routes

import (
	"net/http"

	"github.com/MH-PAVEL/uni-backend-go/internal/handlers"
	"github.com/MH-PAVEL/uni-backend-go/internal/middleware"
)

func RegisterAuthRoutes(mux *http.ServeMux) {
	// Public
	mux.Handle("/api/v1/auth/signup", middleware.Chain(http.HandlerFunc(handlers.Signup)))
	mux.Handle("/api/v1/auth/login", middleware.Chain(http.HandlerFunc(handlers.Login)))
	mux.Handle("/api/v1/auth/logout",  middleware.Chain(http.HandlerFunc(handlers.Logout)))

	//  protected
	mux.Handle("/api/v1/auth/me",
		middleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		}),
			middleware.AuthMiddleware,
		),
	)
}
