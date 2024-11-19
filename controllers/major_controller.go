package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"

	"github.com/google/uuid"
)

// Fetch all majors
func FetchMajors(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	var majors map[string]models.Major
	ref := client.NewRef("majors")
	if err := ref.Get(ctx, &majors); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch majors")
		return
	}

	// Log majors fetched from Firebase
	fmt.Println("Fetched majors:", majors)

	var majorList []models.Major
	for id, major := range majors {
		major.IdMajor = id
		majorList = append(majorList, major)
	}

	utils.RespondJSON(w, http.StatusOK, majorList)
}

// Show a specific major
func ShowMajor(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Id string `json:"idMajor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if requestBody.Id == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "Major ID is required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	var major models.Major
	ref := client.NewRef("majors/" + requestBody.Id)
	if err := ref.Get(ctx, &major); err != nil {
		utils.RespondError(w, http.StatusNotFound, "Major not found")
		return
	}

	major.IdMajor = requestBody.Id
	utils.RespondJSON(w, http.StatusOK, major)
}

// Create a new major
func CreateMajor(w http.ResponseWriter, r *http.Request) {
	var major models.Major
	if err := json.NewDecoder(r.Body).Decode(&major); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if major.TitleMajor == "" || major.IconUrl == "" || major.Link == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "TitleMajor, IconUrl, and Link are required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	id := uuid.New().String()
	ref := client.NewRef("majors/" + id)
	if err := ref.Set(ctx, &major); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create major")
		return
	}

	major.IdMajor = id
	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    major,
		"message": "Major created successfully",
	})
}

// Delete a major
func DeleteMajor(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Id string `json:"idMajor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if requestBody.Id == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "Major ID is required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("majors/" + requestBody.Id)
	if err := ref.Delete(ctx); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to delete major")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Major deleted successfully",
	})
}
