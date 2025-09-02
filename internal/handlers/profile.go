package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/MH-PAVEL/uni-backend-go/internal/database"
	"github.com/MH-PAVEL/uni-backend-go/internal/middleware"
	"github.com/MH-PAVEL/uni-backend-go/internal/models"
	"github.com/MH-PAVEL/uni-backend-go/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProfileCompletionRequest struct {
	FullName            string                    `json:"fullName"            example:"John Doe"`
	Country             string                    `json:"country"             example:"Bangladesh"`
	Address             string                    `json:"address"             example:"Dhaka, Bangladesh"`
	NID                 string                    `json:"nid"                 example:"1234567890123"`
	PlanningMonthToStart string                   `json:"planningMonthToStart" example:"January"`
	PlanningYearToStart  string                   `json:"planningYearToStart"  example:"2025"`
	SSC                 *models.Education         `json:"ssc,omitempty"`
	HSC                 *models.Education         `json:"hsc,omitempty"`
	HigherEducation     *models.HigherEducation   `json:"higherEducation,omitempty"`
	LanguageTests       []models.LanguageTest     `json:"languageTests,omitempty"`
}

type ProfileResponse struct {
	Message string `json:"message" example:"Profile completed successfully"`
}

// @Summary      Complete user profile
// @Description  Complete user profile with additional information after signup
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload  body      ProfileCompletionRequest  true  "Profile completion payload"
// @Success      200      {object}  ProfileResponse
// @Failure      400      {object}  handlers.ErrorResponse  "Invalid request"
// @Failure      401      {object}  handlers.ErrorResponse  "Unauthorized"
// @Failure      409      {object}  handlers.ErrorResponse  "NID already exists"
// @Failure      500      {object}  handlers.ErrorResponse  "Internal error"
// @Router       /api/v1/profile/complete [post]
func CompleteProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Get user ID from context (set by AuthMiddleware)
	userIDStr := r.Context().Value(middleware.CtxUserID)
	if userIDStr == nil {
		utils.ApiError(w, http.StatusUnauthorized, "Missing user id")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		utils.ApiError(w, http.StatusUnauthorized, "Invalid user id")
		return
	}

	var req ProfileCompletionRequest
	if err := utils.SafeDecodeJSON(r, &req); err != nil {
		utils.ApiError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate required fields
	if req.FullName == "" || req.Country == "" || req.Address == "" || req.NID == "" {
		utils.ApiError(w, http.StatusBadRequest, "Full name, country, address, and NID are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	users := database.GetCollection(database.DbName(), database.UsersCollection)

	// Check if NID already exists (unique constraint)
	var existingUser models.User
	err = users.FindOne(ctx, bson.M{"nid": req.NID, "_id": bson.M{"$ne": userID}}).Decode(&existingUser)
	if err == nil {
		utils.ApiError(w, http.StatusConflict, "NID already exists")
		return
	}

	// Update user profile
	updateData := bson.M{
		"$set": bson.M{
			"profileCompletion":     true,
			"fullName":              req.FullName,
			"country":               req.Country,
			"address":               req.Address,
			"nid":                   req.NID,
			"planningMonthToStart":  req.PlanningMonthToStart,
			"planningYearToStart":   req.PlanningYearToStart,
			"ssc":                   req.SSC,
			"hsc":                   req.HSC,
			"higherEducation":       req.HigherEducation,
			"languageTests":         req.LanguageTests,
			"updatedAt":             time.Now(),
		},
	}

	_, err = users.UpdateOne(ctx, bson.M{"_id": userID}, updateData)
	if err != nil {
		utils.ApiError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	utils.ApiResponse(w, http.StatusOK, ProfileResponse{
		Message: "Profile completed successfully",
	})
}

// @Summary      Get user profile
// @Description  Get the current user's profile information
// @Tags         profile
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.User
// @Failure      401  {object}  handlers.ErrorResponse  "Unauthorized"
// @Failure      404  {object}  handlers.ErrorResponse  "User not found"
// @Router       /api/v1/profile [get]
func GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Get user ID from context (set by AuthMiddleware)
	userIDStr := r.Context().Value(middleware.CtxUserID)
	if userIDStr == nil {
		utils.ApiError(w, http.StatusUnauthorized, "Missing user id")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		utils.ApiError(w, http.StatusUnauthorized, "Invalid user id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	users := database.GetCollection(database.DbName(), database.UsersCollection)

	var user models.User
	if err := users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		utils.ApiError(w, http.StatusNotFound, "User not found")
		return
	}

	utils.ApiResponse(w, http.StatusOK, user)
}

// @Summary      Check profile completion status
// @Description  Check if the current user has completed their profile
// @Tags         profile
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Profile completion status"
// @Failure      401  {object}  handlers.ErrorResponse      "Unauthorized"
// @Router       /api/v1/profile/status [get]
func GetProfileStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ApiError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	// Get user ID from context (set by AuthMiddleware)
	userIDStr := r.Context().Value(middleware.CtxUserID)
	if userIDStr == nil {
		utils.ApiError(w, http.StatusUnauthorized, "Missing user id")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		utils.ApiError(w, http.StatusUnauthorized, "Invalid user id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	users := database.GetCollection(database.DbName(), database.UsersCollection)

	var user models.User
	if err := users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		utils.ApiError(w, http.StatusNotFound, "User not found")
		return
	}

	utils.ApiResponse(w, http.StatusOK, map[string]interface{}{
		"profileCompletion": user.ProfileCompletion,
		"hasProfile":        user.ProfileCompletion,
	})
}
