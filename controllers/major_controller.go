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

func ShowMajor(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		TitleMajor string `json:"title_major,omitempty"`
		IdMajor    string `json:"id_major,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Fetch all majors
	var majors map[string]models.Major
	refMajors := client.NewRef("majors")
	if err := refMajors.Get(ctx, &majors); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch majors")
		return
	}

	// Find the matching major
	var matchedMajor models.Major
	var found bool
	var majorId string

	for id, major := range majors {
		if requestBody.IdMajor != "" && id == requestBody.IdMajor {
			matchedMajor = major
			matchedMajor.IdMajor = id
			majorId = id
			found = true
			break
		}
		if requestBody.TitleMajor != "" && major.TitleMajor == requestBody.TitleMajor {
			matchedMajor = major
			matchedMajor.IdMajor = id
			majorId = id
			found = true
			break
		}
	}

	if !found {
		utils.RespondError(w, http.StatusNotFound, "Major not found")
		return
	}

	// Fetch all categories related to the major
	var categories map[string]models.Category
	refCategories := client.NewRef("categories")
	if err := refCategories.Get(ctx, &categories); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	var relatedCategories []map[string]interface{}
	for categoryId, category := range categories {
		if category.IdMajor == majorId {
			// Fetch all services related to the current category
			var services map[string]models.Service
			refServices := client.NewRef("services")
			if err := refServices.Get(ctx, &services); err != nil {
				utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch services")
				return
			}

			var relatedServices []models.Service
			for _, service := range services {
				if service.IdCategory == categoryId {
					relatedServices = append(relatedServices, service)
				}
			}

			// Append category and its services
			relatedCategories = append(relatedCategories, map[string]interface{}{
				"category": category,
				"services": relatedServices,
			})
		}
	}

	// Respond with major, categories, and services
	response := map[string]interface{}{
		"major":      matchedMajor,
		"categories": relatedCategories,
	}

	utils.RespondJSON(w, http.StatusOK, response)
}

// Create a new major
func CreateMajor(w http.ResponseWriter, r *http.Request) {
	var major models.Major
	if err := json.NewDecoder(r.Body).Decode(&major); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if major.TitleMajor == "" || major.IconUrl == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "TitleMajor and IconUrl are required")
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
