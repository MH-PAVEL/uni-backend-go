package handlers

import (
	"fmt"
	"net/http"
)

// HealthCheck is for server monitoring
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "✅ Server is healthy")
}
