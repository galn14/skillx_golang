package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"fmt"
	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"

	"github.com/google/uuid"
)

// Fetch all services
func FetchServices(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	var services map[string]models.Service
	ref := client.NewRef("services")
	if err := ref.Get(ctx, &services); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch services")
		return
	}

	var serviceList []models.Service
	for id, service := range services {
		service.IdService = id
		serviceList = append(serviceList, service)
	}

	utils.RespondJSON(w, http.StatusOK, serviceList)
}

func ShowService(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Id           string `json:"id,omitempty"`
		TitleService string `json:"title_service,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	fmt.Println("Request Body:", requestBody) // Log input JSON

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Ambil semua layanan
	var services map[string]models.Service
	ref := client.NewRef("services")
	if err := ref.Get(ctx, &services); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch services")
		return
	}

	fmt.Println("Fetched Services:", services) // Log data layanan yang diambil

	// Cari layanan berdasarkan ID atau TitleService
	for id, service := range services {
		if requestBody.Id != "" && id == requestBody.Id {
			service.IdService = id
			utils.RespondJSON(w, http.StatusOK, service)
			return
		}
		if requestBody.TitleService != "" && service.TitleService == requestBody.TitleService {
			service.IdService = id
			utils.RespondJSON(w, http.StatusOK, service)
			return
		}
	}

	utils.RespondError(w, http.StatusNotFound, "Service not found")
}

func CreateService(w http.ResponseWriter, r *http.Request) {
	var serviceInput struct {
		TitleService  string `json:"title_service"`
		IconUrl       string `json:"icon_url"`
		TitleCategory string `json:"title_category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&serviceInput); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	// Validasi input
	if serviceInput.TitleService == "" || serviceInput.IconUrl == "" || serviceInput.TitleCategory == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "TitleService, IconUrl, and TitleCategory are required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Cari IdCategory berdasarkan TitleCategory
	var categories map[string]models.Category
	refCategories := client.NewRef("categories")
	if err := refCategories.Get(ctx, &categories); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	var idCategory string
	for id, category := range categories {
		if category.Title == serviceInput.TitleCategory {
			idCategory = id
			break
		}
	}

	if idCategory == "" {
		utils.RespondError(w, http.StatusBadRequest, "TitleCategory not found")
		return
	}

	// Buat service
	service := models.Service{
		TitleService: serviceInput.TitleService,
		IconUrl:      serviceInput.IconUrl,
		IdCategory:   idCategory,
	}

	id := uuid.New().String()
	ref := client.NewRef("services/" + id)
	if err := ref.Set(ctx, &service); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create service")
		return
	}

	service.IdService = id
	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    service,
		"message": "Service created successfully",
	})
}

// Delete a service
func DeleteService(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Id string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if requestBody.Id == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "Service ID is required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("services/" + requestBody.Id)
	if err := ref.Delete(ctx); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to delete service")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Service deleted successfully",
	})
}

// Update an existing service
func UpdateService(w http.ResponseWriter, r *http.Request) {
	// Get id_service from query parameters
	idService := r.URL.Query().Get("id_service")
	if idService == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "Service ID is required")
		return
	}

	var requestBody struct {
		TitleService  string `json:"title_service,omitempty"`
		IconUrl       string `json:"icon_url,omitempty"`
		TitleCategory string `json:"title_category,omitempty"`
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

	// Reference to the specific service
	ref := client.NewRef("services/" + idService)

	// Check if the service exists
	var existingService models.Service
	if err := ref.Get(ctx, &existingService); err != nil {
		utils.RespondError(w, http.StatusNotFound, "Service not found")
		return
	}

	// Update fields if provided
	if requestBody.TitleService != "" {
		existingService.TitleService = requestBody.TitleService
	}
	if requestBody.IconUrl != "" {
		existingService.IconUrl = requestBody.IconUrl
	}

	// Update category if TitleCategory is provided
	if requestBody.TitleCategory != "" {
		// Fetch categories to find the corresponding ID
		var categories map[string]models.Category
		refCategories := client.NewRef("categories")
		if err := refCategories.Get(ctx, &categories); err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch categories")
			return
		}

		var idCategory string
		for id, category := range categories {
			if category.Title == requestBody.TitleCategory {
				idCategory = id
				break
			}
		}

		if idCategory == "" {
			utils.RespondError(w, http.StatusBadRequest, "TitleCategory not found")
			return
		}

		existingService.IdCategory = idCategory
	}

	// Save updated service back to Firebase
	if err := ref.Set(ctx, &existingService); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to update service")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    existingService,
		"message": "Service updated successfully",
	})
}
