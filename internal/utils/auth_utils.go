package utils

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MH-PAVEL/uni-backend-go/internal/config"
	"github.com/MH-PAVEL/uni-backend-go/internal/database"
	"github.com/MH-PAVEL/uni-backend-go/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DbName returns the database name from config
func DbName() string { 
	cfg := config.AppConfig
	if cfg == nil {
		return ""
	}
	return cfg.Database.Name 
}

// Collection names
func UsersCollection() string { return "users" }
func RefreshTokensCollection() string { return "refresh_tokens" }

// IssueTokens creates a short-lived access JWT and a long-lived refresh token (hashed & stored).
func IssueTokens(ctx context.Context, userID primitive.ObjectID) (access, refresh string, err error) {
	cfg := config.AppConfig
	if cfg == nil {
		return "", "", fmt.Errorf("configuration not loaded")
	}

	access, err = GenerateJWT(cfg.Auth.JWTSecret, userID.Hex(), cfg.Auth.AccessTTL)
	if err != nil {
		return "", "", err
	}

	refresh, err = GenerateSecureToken(32)
	if err != nil {
		return "", "", err
	}
	hash := SHA256Hex(refresh)

	now := time.Now()
	doc := models.RefreshToken{
		UserID:    userID,
		TokenHash: hash,
		CreatedAt: now,
		ExpiresAt: now.Add(cfg.Auth.RefreshTTL),
	}

	col := database.GetCollection(DbName(), RefreshTokensCollection())
	if _, err := col.InsertOne(ctx, doc); err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

// SetRefreshCookie sets the refresh token as HttpOnly cookie.
// For cross-site production use: SameSite=None and Secure=true.
func SetRefreshCookie(w http.ResponseWriter, token string) {
	cfg := config.AppConfig
	if cfg == nil {
		return
	}
	
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		Path:     "/api/v1/auth",
		MaxAge:   int(cfg.Auth.RefreshTTL.Seconds()),
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// GetRefreshTokenFromReq reads refresh token from cookie first, then JSON body (Postman fallback).
func GetRefreshTokenFromReq(r *http.Request) string {
	if c, err := r.Cookie("refresh_token"); err == nil && c.Value != "" {
		return c.Value
	}
	var rr RefreshRequest
	_ = SafeDecodeJSON(r, &rr)
	return strings.TrimSpace(rr.RefreshToken)
}
