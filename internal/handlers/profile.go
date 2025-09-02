package handlers

import (
	"context"
	"fmt"
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
	FullName            string
	Country             string
	Address             string
	NID                 string
	PlanningMonthToStart string
	PlanningYearToStart  string
	SSC                 *models.Education
	HSC                 *models.HigherEducation
	HigherEducation     *models.HigherEducation
	LanguageTests       []models.LanguageTest
}

type ProfileResponse struct {
	Message string `json:"message" example:"Profile completed successfully"`
}

// @Summary      Complete user profile
// @Description  Complete user profile with additional information and document uploads after signup
// @Tags         profile
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        fullName formData string true "Full name"
// @Param        country formData string true "Country"
// @Param        address formData string true "Address"
// @Param        nid formData string true "National ID"
// @Param        planningMonthToStart formData string false "Planning month to start"
// @Param        planningYearToStart formData string false "Planning year to start"
// @Param        sscCertificate formData file true "SSC Certificate (PDF)"
// @Param        sscMarksheet formData file false "SSC Marksheet (PDF, optional)"
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
	// // userIDStr := r.Context().Value(middleware.CtxUserID)
	// userIDStr := "68b62739cc8f4985cff260b0"
	// if userIDStr == nil {
	// 	utils.ApiError(w, http.StatusUnauthorized, "Missing user id")
	// 	return
	// }

	// userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	// if err != nil {
	// 	utils.ApiError(w, http.StatusUnauthorized, "Invalid user id")
	// 	return
	// }

	userID := "68b7666f99daa59355222735"

	// Parse multipart form data (max 32MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		utils.ApiError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Extract form fields
	req := ProfileCompletionRequest{
		FullName:             r.FormValue("fullName"),
		Country:              r.FormValue("country"),
		Address:              r.FormValue("address"),
		NID:                  r.FormValue("nid"),
		PlanningMonthToStart: r.FormValue("planningMonthToStart"),
		PlanningYearToStart:  r.FormValue("planningYearToStart"),
	}

	// Validate required fields
	if req.FullName == "" || req.Country == "" || req.Address == "" || req.NID == "" {
		utils.ApiError(w, http.StatusBadRequest, "Full name, country, address, and NID are required")
		return
	}

	// Get uploaded files
	sscCertificateFile, sscCertificateHeader, err := r.FormFile("sscCertificate")
	if err != nil {
		utils.ApiError(w, http.StatusBadRequest, "SSC Certificate is required")
		return
	}
	defer sscCertificateFile.Close()

	sscMarksheetFile, sscMarksheetHeader, err := r.FormFile("sscMarksheet")
	var sscMarksheetUploaded map[string]interface{}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Upload SSC Certificate (required)
	sscCertificateUploaded, err := utils.UploadFile(ctx, sscCertificateHeader, "ssc_certificates")
	if err != nil {
		utils.ApiError(w, http.StatusBadRequest, fmt.Sprintf("Failed to upload SSC Certificate: %v", err))
		return
	}

	// Upload SSC Marksheet (optional)
	if sscMarksheetFile != nil && sscMarksheetHeader != nil {
		sscMarksheetUploaded, err = utils.UploadFile(ctx, sscMarksheetHeader, "ssc_marksheets")
		if err != nil {
			utils.ApiError(w, http.StatusBadRequest, fmt.Sprintf("Failed to upload SSC Marksheet: %v", err))
			return
		}
	}

	users := database.GetCollection(database.DbName(), database.UsersCollection)

	// Check if NID already exists (unique constraint)
	var existingUser models.User
	err = users.FindOne(ctx, bson.M{"nid": req.NID, "_id": bson.M{"$ne": userID}}).Decode(&existingUser)
	if err == nil {
		utils.ApiError(w, http.StatusConflict, "NID already exists")
		return
	}

	// Prepare document data
	sscCertificateDoc := &models.Document{
		FileID:   sscCertificateUploaded["fileId"].(string),
		FileName: sscCertificateUploaded["fileName"].(string),
		FileURL:  sscCertificateUploaded["fileUrl"].(string),
		FileSize: sscCertificateUploaded["fileSize"].(int64),
		MimeType: sscCertificateUploaded["mimeType"].(string),
	}

	var sscMarksheetDoc *models.Document
	if sscMarksheetUploaded != nil {
		sscMarksheetDoc = &models.Document{
			FileID:   sscMarksheetUploaded["fileId"].(string),
			FileName: sscMarksheetUploaded["fileName"].(string),
			FileURL:  sscMarksheetUploaded["fileUrl"].(string),
			FileSize: sscMarksheetUploaded["fileSize"].(int64),
			MimeType: sscMarksheetUploaded["mimeType"].(string),
		}
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
			"ssCertificate":         sscCertificateDoc,
			"sscMarksheet":          sscMarksheetDoc,
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
