package routes

import (
	"net/http"

	"github.com/MH-PAVEL/uni-backend-go/internal/handlers"
	"github.com/MH-PAVEL/uni-backend-go/internal/middleware"
)

func RegisterProfileRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/profile/complete",
		middleware.Chain(
			http.HandlerFunc(handlers.CompleteProfile),
			middleware.AuthMiddleware,
		),
	)

	mux.Handle("GET /api/v1/profile",
		middleware.Chain(
			http.HandlerFunc(handlers.GetProfile),
			middleware.AuthMiddleware,
		),
	)

	mux.Handle("GET /api/v1/profile/status",
		middleware.Chain(
			http.HandlerFunc(handlers.GetProfileStatus),
			middleware.AuthMiddleware,
		),
	)
}
