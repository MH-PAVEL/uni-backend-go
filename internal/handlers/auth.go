package handlers

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/MH-PAVEL/uni-backend-go/internal/config"
	"github.com/MH-PAVEL/uni-backend-go/internal/database"
	"github.com/MH-PAVEL/uni-backend-go/internal/middleware"
	"github.com/MH-PAVEL/uni-backend-go/internal/models"
	"github.com/MH-PAVEL/uni-backend-go/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type SignupRequest struct {
	Email    string `json:"email" example:"F2HbU@example.com"`
	Phone    string `json:"phone" example:"01234567890"`
	Password string `json:"password" example:"password123"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" example:"F2HbU@example.com"` // email or phone
	Password   string `json:"password" example:"password123"`
}

type AuthResponse struct {
	Token        string      `json:"token" example:"<jwt_access_token>"`
	RefreshToken string      `json:"refreshToken" example:"<refresh_token>"`
	User         interface{} `json:"user,omitempty"`
}

// LogoutResponse documents a simple message payload
type LogoutResponse struct {
    Message string `json:"message" example:"logged out"`
}


// ErrorResponse is a generic error schema
// @Description Error response with a message field
type ErrorResponse struct {
    Message string `json:"message" example:"Unauthorized"`
}


// @Summary      Signup
// @Description  Create a user and return access & refresh tokens (also set as cookies).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body      SignupRequest  true  "Signup payload"
// @Success      200      {object}  AuthResponse
// @Failure      400      {object}  ErrorResponse  "Invalid request"
// @Failure      409      {object}  ErrorResponse  "User already exists"
// @Failure      500      {object}  ErrorResponse  "Internal error"
// @Router       /api/v1/auth/signup [post]
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

	users := database.GetCollection(database.DbName(), database.UsersCollection)

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

	access, refresh, err := utils.IssueTokens(ctx, uid)
	if err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Failed to generate authentication tokens")
		return
	}

	// set cookies
	utils.SetAccessCookie(w, access)
	utils.SetRefreshCookie(w, refresh)

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

// @Summary      Login
// @Description  Login with email or phone and get access & refresh tokens (also set as cookies).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body      LoginRequest  true  "Login payload"
// @Success      200      {object}  AuthResponse
// @Failure      400      {object}  ErrorResponse  "Invalid request"
// @Failure      401      {object}  ErrorResponse  "Invalid credentials"
// @Failure      500      {object}  ErrorResponse  "Internal error"
// @Router       /api/v1/auth/login [post]
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

	col := database.GetCollection(database.DbName(), database.UsersCollection)
	filter := bson.M{
		"$or": []bson.M{
			{"email": req.Identifier},
			{"phone": req.Identifier},
		},
	}

	var user bson.M
	if err := col.FindOne(ctx, filter).Decode(&user); err != nil {
		utils.ApiError(w, http.StatusUnauthorized, "User not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user["password"].(string)), []byte(req.Password)); err != nil {
		utils.ApiError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	uid := user["_id"].(primitive.ObjectID)

	access, refresh, err := utils.IssueTokens(ctx, uid)
	if err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Failed to generate authentication tokens")
		return
	}
	utils.SetAccessCookie(w, access)
	utils.SetRefreshCookie(w, refresh)

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

// @Summary      Refresh access token
// @Description  Rotate refresh token and return a new access token (cookies updated). Requires valid refresh token cookie.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  AuthResponse
// @Failure      400  {object}  ErrorResponse  "Missing refresh token"
// @Failure      401  {object}  ErrorResponse  "Invalid or expired refresh token"
// @Failure      500  {object}  ErrorResponse  "Internal error"
// @Router       /api/v1/auth/refresh [post]
func Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}
	token := utils.GetRefreshTokenFromReq(r)
	if token == "" {
		utils.ApiError(w, http.StatusBadRequest, "Missing refresh token")
		return
	}
	hash := utils.SHA256Hex(token)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rtc := database.GetCollection(database.DbName(), database.RefreshTokensCollection)

	var rt models.RefreshToken
	if err := rtc.FindOne(ctx, bson.M{"tokenHash": hash}).Decode(&rt); err != nil {
		utils.ApiError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}
	if rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) {
		utils.ApiError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	// Rotate refresh token
	now := time.Now()
	newRefresh, err := utils.GenerateSecureToken(32)
	if err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}
	newHash := utils.SHA256Hex(newRefresh)

	if _, err := rtc.UpdateByID(ctx, rt.ID, bson.M{
		"$set": bson.M{
			"revokedAt":           now,
			"replacedByTokenHash": newHash,
		},
	}); err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Failed to revoke old refresh token")
		return
	}

	cfg := config.AppConfig
	if cfg == nil {
		utils.ApiError(w, http.StatusInternalServerError, "Configuration not loaded")
		return
	}

	_, err = rtc.InsertOne(ctx, models.RefreshToken{
		UserID:    rt.UserID,
		TokenHash: newHash,
		CreatedAt: now,
		ExpiresAt: now.Add(cfg.Auth.RefreshTTL),
	})
	if err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Failed to store new refresh token")
		return
	}

	// New access token
	access, err := utils.GenerateJWT(cfg.Auth.JWTSecret, rt.UserID.Hex(), cfg.Auth.AccessTTL)
	if err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}
	
	utils.SetAccessCookie(w, access)
	utils.SetRefreshCookie(w, newRefresh)

	utils.ApiResponse(w, http.StatusOK, AuthResponse{
		Token:        access,
		RefreshToken: newRefresh,
	})
}

// @Summary      Logout
// @Description  Revoke current refresh token and clear auth cookies.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  LogoutResponse
// @Failure      400  {object}  ErrorResponse  "Missing refresh token"
// @Router       /api/v1/auth/logout [post]
func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}
	token := utils.GetRefreshTokenFromReq(r)
	if token == "" {
		utils.ApiError(w, http.StatusBadRequest, "Missing refresh token")
		return
	}
	hash := utils.SHA256Hex(token)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rtc := database.GetCollection(database.DbName(), database.RefreshTokensCollection)
	now := time.Now()
	_, _ = rtc.UpdateOne(ctx, bson.M{
		"tokenHash": hash, "revokedAt": bson.M{"$exists": false},
	}, bson.M{
		"$set": bson.M{"revokedAt": now},
	})


	// Clear refresh cookie
	http.SetCookie(w, &http.Cookie{
		Name:     utils.RefreshTokenCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Clear access cookie
	http.SetCookie(w, &http.Cookie{
		Name:     utils.AccessTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})


	utils.ApiResponse(w, http.StatusOK, LogoutResponse{
		Message: "Logout successful",
	})
}

// @Summary      Get current user
// @Description  Return the authenticated user.
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.User
// @Failure      401  {object}  ErrorResponse  "Unauthorized"
// @Failure      404  {object}  ErrorResponse  "User not found"
// @Router       /api/v1/auth/me [get]
func GetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}
	uid := r.Context().Value(middleware.CtxUserID)
	if uid == nil {
		utils.ApiError(w, http.StatusUnauthorized, "Missing user id")
		return
	}

	// Convert userID string -> ObjectID
	userID, err := primitive.ObjectIDFromHex(uid.(string))
	if err != nil {
		utils.ApiError(w, http.StatusUnauthorized, "Invalid user id")
		return
	}
	
	// DB context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	col := database.GetCollection(database.DbName(), database.UsersCollection)

	var user models.User
	if err := col.FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		utils.ApiError(w, http.StatusNotFound, "User not found")
		return
	}

	utils.ApiResponse(w, http.StatusOK, user)
}
