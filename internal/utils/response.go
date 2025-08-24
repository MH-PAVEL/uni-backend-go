package utils

import (
	"encoding/json"
	"net/http"
)

func ApiResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func ApiError(w http.ResponseWriter, status int, msg string) {
	ApiResponse(w, status, map[string]string{"error": msg})
}

// Optional: strict decoder helper
func SafeDecodeJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
