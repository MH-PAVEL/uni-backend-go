package handlers

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MH-PAVEL/uni-backend-go/internal/config"
	"github.com/MH-PAVEL/uni-backend-go/internal/database"
	"github.com/MH-PAVEL/uni-backend-go/internal/models"
	"github.com/MH-PAVEL/uni-backend-go/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	AccessTTL  = 7 * 24 * time.Hour
	RefreshTTL = 30 * 24 * time.Hour
)

func DbName() string { return os.Getenv("MONGO_DB_NAME") }
func usersColl() string { return "users" }
func rtColl() string    { return "refresh_tokens" }

// issueTokens creates a short-lived access JWT and a long-lived refresh token (hashed & stored).
func issueTokens(ctx context.Context, userID primitive.ObjectID) (access, refresh string, err error) {
	secret := config.Get("JWT_SECRET")

	access, err = utils.GenerateJWT(secret, userID.Hex(), AccessTTL)
	if err != nil {
		return "", "", err
	}

	refresh, err = utils.GenerateSecureToken(32)
	if err != nil {
		return "", "", err
	}
	hash := utils.SHA256Hex(refresh)

	now := time.Now()
	doc := models.RefreshToken{
		UserID:    userID,
		TokenHash: hash,
		CreatedAt: now,
		ExpiresAt: now.Add(RefreshTTL),
	}

	col := database.GetCollection(DbName(), rtColl())
	if _, err := col.InsertOne(ctx, doc); err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

// setRefreshCookie sets the refresh token as HttpOnly cookie.
// For cross-site production use: SameSite=None and Secure=true.
func setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		Path:     "/api/v1/auth",
		MaxAge:   int(RefreshTTL.Seconds()),
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// getRefreshTokenFromReq reads refresh token from cookie first, then JSON body (Postman fallback).
func getRefreshTokenFromReq(r *http.Request) string {
	if c, err := r.Cookie("refresh_token"); err == nil && c.Value != "" {
		return c.Value
	}
	var rr RefreshRequest
	_ = utils.SafeDecodeJSON(r, &rr)
	return strings.TrimSpace(rr.RefreshToken)
}
