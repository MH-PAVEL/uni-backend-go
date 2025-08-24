package routes

import (
	"net/http"

	_ "github.com/MH-PAVEL/uni-backend-go/internal/docs"
	"github.com/MH-PAVEL/uni-backend-go/internal/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

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
