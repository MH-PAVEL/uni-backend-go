package routes

import (
	"net/http"

	"github.com/MH-PAVEL/uni-backend-go/internal/handlers"
)

func RegisterHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", handlers.HealthCheck)
}