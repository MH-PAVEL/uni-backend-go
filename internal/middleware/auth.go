package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/MH-PAVEL/uni-backend-go/internal/config"
	"github.com/MH-PAVEL/uni-backend-go/internal/utils"
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

	unauth := func(w http.ResponseWriter) {
		w.Header().Set("WWW-Authenticate", `Bearer realm="api"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	extractToken := func(r *http.Request) string {
		// 1) Prefer Authorization header: Bearer <token> (case-insensitive)
		h := strings.TrimSpace(r.Header.Get("Authorization"))
		if h != "" {
			parts := strings.Fields(h)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				return parts[1]
			}
		}
		// 2) Fallback to access_token cookie
		if c, err := r.Cookie(utils.AccessTokenCookieName); err == nil && c.Value != "" {
			return c.Value
		}
		return ""
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := extractToken(r)
		if tokenStr == "" {
			unauth(w)
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			// Enforce HS* family
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(cfg.Auth.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			unauth(w)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			unauth(w)
			return
		}

		// Extract sub robustly (string or number)
		var sub string
		switch v := claims["sub"].(type) {
		case string:
			sub = v
		case float64:
			sub = strconv.FormatInt(int64(v), 10)
		default:
			unauth(w)
			return
		}
		if sub == "" {
			unauth(w)
			return
		}

		ctx := context.WithValue(r.Context(), CtxUserID, sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
