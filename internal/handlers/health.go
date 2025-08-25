package handlers

import (
	"net/http"

	"github.com/MH-PAVEL/uni-backend-go/internal/utils"
)

// HealthCheck is for server monitoring
// @Summary      Health check
// @Description  Returns 200 OK when the server is healthy
// @Tags         health
// @Produce      plain
// @Success      200 {string} string "✅ Server is healthy"
// @Router       /api/health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.ApiResponse(w, http.StatusOK,  map[string]string{"message": "✅ Server is healthy"})
}
