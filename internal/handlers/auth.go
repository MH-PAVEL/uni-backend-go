package handlers

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/MH-PAVEL/uni-backend-go/internal/config"
	"github.com/MH-PAVEL/uni-backend-go/internal/database"
	"github.com/MH-PAVEL/uni-backend-go/internal/models"
	"github.com/MH-PAVEL/uni-backend-go/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type SignupRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"` // email or phone
	Password   string `json:"password"`
}

type AuthResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refreshToken"`
	User         interface{} `json:"user,omitempty"`
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	var req SignupRequest
	if err := utils.SafeDecodeJSON(r, &req); err != nil {
		utils.ApiError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)

	if !emailRegex.MatchString(req.Email) {
		utils.ApiError(w, http.StatusBadRequest, "Invalid email format")
		return
	}
	if len(req.Phone) != 11 {
		utils.ApiError(w, http.StatusBadRequest, "Phone number must be 11 digits")
		return
	}
	if len(req.Password) == 0 {
		utils.ApiError(w, http.StatusBadRequest, "Missing password")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	users := database.GetCollection(DbName(), usersColl())

	// Uniqueness check
	filter := bson.M{"$or": []bson.M{{"email": req.Email}, {"phone": req.Phone}}}
	var ex bson.M
	if err := users.FindOne(ctx, filter).Decode(&ex); err == nil {
		utils.ApiError(w, http.StatusBadRequest, "User already exists")
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}

	now := time.Now()
	doc := bson.M{
		"email":     req.Email,
		"phone":     req.Phone,
		"password":  string(hash),
		"createdAt": now,
		"updatedAt": now,
	}
	res, err := users.InsertOne(ctx, doc)
	if err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Could not create user")
		return
	}
	uid := res.InsertedID.(primitive.ObjectID)

	access, refresh, err := issueTokens(ctx, uid)
	if err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Could not issue tokens")
		return
	}
	setRefreshCookie(w, refresh)

	utils.ApiResponse(w, http.StatusOK, AuthResponse{
		Token:        access,
		RefreshToken: refresh,
		User: bson.M{
			"id":        uid.Hex(),
			"email":     req.Email,
			"phone":     req.Phone,
			"createdAt": now,
		},
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	var req LoginRequest
	if err := utils.SafeDecodeJSON(r, &req); err != nil {
		utils.ApiError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	req.Identifier = strings.TrimSpace(strings.ToLower(req.Identifier))

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	col := database.GetCollection(DbName(), usersColl())
	filter := bson.M{
		"$or": []bson.M{
			{"email": req.Identifier},
			{"phone": req.Identifier},
		},
	}

	var user bson.M
	if err := col.FindOne(ctx, filter).Decode(&user); err != nil {
		utils.ApiError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user["password"].(string)), []byte(req.Password)); err != nil {
		utils.ApiError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	uid := user["_id"].(primitive.ObjectID)

	access, refresh, err := issueTokens(ctx, uid)
	if err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Could not issue tokens")
		return
	}
	setRefreshCookie(w, refresh)

	utils.ApiResponse(w, http.StatusOK, AuthResponse{
		Token:        access,
		RefreshToken: refresh,
		User: bson.M{
			"id":        uid.Hex(),
			"email":     user["email"],
			"phone":     user["phone"],
			"createdAt": user["createdAt"],
		},
	})
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}
	token := getRefreshTokenFromReq(r)
	if token == "" {
		utils.ApiError(w, http.StatusBadRequest, "Missing refresh token")
		return
	}
	hash := utils.SHA256Hex(token)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rtc := database.GetCollection(DbName(), rtColl())

	var rt models.RefreshToken
	if err := rtc.FindOne(ctx, bson.M{"tokenHash": hash}).Decode(&rt); err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Could not issue tokens")
		return
	}
	if rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) {
		utils.ApiError(w, http.StatusInternalServerError, "Could not issue tokens")
		return
	}

	// Rotate refresh token
	now := time.Now()
	newRefresh, err := utils.GenerateSecureToken(32)
	if err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Could not issue tokens")
		return
	}
	newHash := utils.SHA256Hex(newRefresh)

	if _, err := rtc.UpdateByID(ctx, rt.ID, bson.M{
		"$set": bson.M{
			"revokedAt":           now,
			"replacedByTokenHash": newHash,
		},
	}); err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Could not Refresh issue tokens")
		return
	}

	_, err = rtc.InsertOne(ctx, models.RefreshToken{
		UserID:    rt.UserID,
		TokenHash: newHash,
		CreatedAt: now,
		ExpiresAt: now.Add(RefreshTTL),
	})
	if err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Could not issue tokens")
		return
	}

	// New access token
	secret := config.Get("JWT_SECRET")
	access, err := utils.GenerateJWT(secret, rt.UserID.Hex(), AccessTTL)
	if err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Could not issue tokens")
		return
	}
	setRefreshCookie(w, newRefresh)

	utils.ApiResponse(w, http.StatusOK, AuthResponse{
		Token:        access,
		RefreshToken: newRefresh,
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}
	token := getRefreshTokenFromReq(r)
	if token == "" {
		utils.ApiError(w, http.StatusBadRequest, "Missing refresh token")
		return
	}
	hash := utils.SHA256Hex(token)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rtc := database.GetCollection(DbName(), rtColl())
	now := time.Now()
	_, _ = rtc.UpdateOne(ctx, bson.M{
		"tokenHash": hash, "revokedAt": bson.M{"$exists": false},
	}, bson.M{
		"$set": bson.M{"revokedAt": now},
	})

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		HttpOnly: true,
	})

	utils.ApiResponse(w, http.StatusOK, map[string]string{"message": "logged out"})
}
