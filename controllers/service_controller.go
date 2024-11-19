package controllers

import (
	"context"
	"encoding/json"
	"net/http"

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

// Show a specific service
func ShowService(w http.ResponseWriter, r *http.Request) {
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

	var service models.Service
	ref := client.NewRef("services/" + requestBody.Id)
	if err := ref.Get(ctx, &service); err != nil {
		utils.RespondError(w, http.StatusNotFound, "Service not found")
		return
	}

	service.IdService = requestBody.Id
	utils.RespondJSON(w, http.StatusOK, service)
}

// Create a new service
func CreateService(w http.ResponseWriter, r *http.Request) {
	var service models.Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if service.TitleService == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "TitleService is required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
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
