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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"
)

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// IssueTokens creates a short-lived access JWT and a long-lived refresh token
func IssueTokens(ctx context.Context, userID primitive.ObjectID) (access, refresh string, err error) {
	cfg := config.AppConfig
	if cfg == nil {
		return "", "", fmt.Errorf("configuration not loaded")
	}

	// Generate access token
	access, err = GenerateJWT(cfg.Auth.JWTSecret, userID.Hex(), cfg.Auth.AccessTTL)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refresh, err = GenerateSecureToken(32)
	if err != nil {
		return "", "", err
	}

	// Hash refresh token and store in user document
	hash := SHA256Hex(refresh)
	expiresAt := time.Now().Add(cfg.Auth.RefreshTTL)

	users := database.GetCollection(database.DbName(), database.UsersCollection)
	_, err = users.UpdateOne(ctx, 
		bson.M{"_id": userID},
		bson.M{
			"$set": bson.M{
				"refreshTokenHash": hash,
				"refreshTokenExpires": expiresAt,
				"updatedAt": time.Now(),
			},
		},
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return access, refresh, nil
}

// ValidateRefreshToken validates a refresh token against the user's stored hash
func ValidateRefreshToken(ctx context.Context, userID primitive.ObjectID, token string) (bool, error) {
	users := database.GetCollection(database.DbName(), database.UsersCollection)
	
	var user models.User
	err := users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return false, err
	}

	// Check if token exists and is not expired
	if user.RefreshTokenHash == "" || user.RefreshTokenExpires == nil {
		return false, nil
	}

	if time.Now().After(*user.RefreshTokenExpires) {
		return false, nil
	}

	// Verify token hash
	tokenHash := SHA256Hex(token)
	return tokenHash == user.RefreshTokenHash, nil
}

// RevokeRefreshToken removes the refresh token from the user
func RevokeRefreshToken(ctx context.Context, userID primitive.ObjectID) error {
	users := database.GetCollection(database.DbName(), database.UsersCollection)
	
	_, err := users.UpdateOne(ctx,
		bson.M{"_id": userID},
		bson.M{
			"$unset": bson.M{
				"refreshTokenHash": "",
				"refreshTokenExpires": "",
			},
			"$set": bson.M{
				"updatedAt": time.Now(),
			},
		},
	)
	return err
}

// SetRefreshCookie sets the refresh token as HttpOnly cookie.
// For cross-site production use: SameSite=None and Secure=true.
func SetRefreshCookie(w http.ResponseWriter, token string) {
	cfg := config.AppConfig
	if cfg == nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		Path:     "/api/v1/auth",
		MaxAge:   int(cfg.Auth.RefreshTTL.Seconds()),
	})
}

// GetRefreshTokenFromReq reads refresh token from cookie first, then JSON body
func GetRefreshTokenFromReq(r *http.Request) string {
	if c, err := r.Cookie(RefreshTokenCookieName); err == nil && c.Value != "" {
		return c.Value
	}
	var rr RefreshRequest
	_ = SafeDecodeJSON(r, &rr)
	return strings.TrimSpace(rr.RefreshToken)
}

// SetAccessCookie sets the access token as HttpOnly cookie
func SetAccessCookie(w http.ResponseWriter, token string) {
	cfg := config.AppConfig
	if cfg == nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookieName,
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		Path:     "/",
		MaxAge:   int(cfg.Auth.AccessTTL.Seconds()),
	})
}