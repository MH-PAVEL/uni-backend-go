package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/MH-PAVEL/uni-backend-go/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const CtxUserID ctxKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	cfg := config.AppConfig
	if cfg == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(h, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(cfg.Auth.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		sub, _ := claims["sub"].(string)

		ctx := context.WithValue(r.Context(), CtxUserID, sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
