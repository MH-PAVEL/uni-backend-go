package routes

import (
	"net/http"

	"github.com/MH-PAVEL/uni-backend-go/internal/middleware"
)

func RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Attach different route groups
	RegisterHealthRoutes(mux)
	RegisterAuthRoutes(mux)
	// RegisterUserRoutes(mux)

	// Applied global middlewares (CORS, Logger, RateLimiter)
	handler := middleware.Chain(mux,
		middleware.CORSMiddleware,
		middleware.LoggerMiddleware,
		middleware.RateLimiterMiddleware,
	)

	return handler
}
