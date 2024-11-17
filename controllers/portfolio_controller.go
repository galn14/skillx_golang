package controllers

import (
	"context"
	"encoding/json"
	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"
	"net/http"

	"github.com/google/uuid"
)

// ListPortfolios - GET /user/portfolios/view
func ListPortfolios(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value("uid").(string)
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase database")
		return
	}

	ref := client.NewRef("portfolios/" + uid)
	var portfolios []models.Portfolio
	err = ref.Get(context.Background(), &portfolios)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch portfolios")
		return
	}

	utils.RespondJSON(w, http.StatusOK, portfolios)
}

// CreatePortfolio - POST /user/portfolios/create
func CreatePortfolio(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value("uid").(string)
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	var portfolio models.Portfolio
	if err := json.NewDecoder(r.Body).Decode(&portfolio); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	portfolio.ID = uuid.New().String()
	portfolio.UserID = uid

	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase database")
		return
	}

	ref := client.NewRef("portfolios/" + uid + "/" + portfolio.ID)
	err = ref.Set(context.Background(), portfolio)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create portfolio")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, portfolio)
}

// ShowPortfolio - GET /user/portfolios/view/{id}
func ShowPortfolio(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value("uid").(string)
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, http.StatusBadRequest, "Portfolio ID is required")
		return
	}

	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase database")
		return
	}

	ref := client.NewRef("portfolios/" + uid + "/" + id)
	var portfolio models.Portfolio
	err = ref.Get(context.Background(), &portfolio)
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, "Portfolio not found")
		return
	}

	utils.RespondJSON(w, http.StatusOK, portfolio)
}

// UpdatePortfolio - PUT /user/portfolios/update/{id}
func UpdatePortfolio(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value("uid").(string)
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, http.StatusBadRequest, "Portfolio ID is required")
		return
	}

	var updatedPortfolio models.Portfolio
	if err := json.NewDecoder(r.Body).Decode(&updatedPortfolio); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase database")
		return
	}

	ref := client.NewRef("portfolios/" + uid + "/" + id)
	err = ref.Set(context.Background(), updatedPortfolio)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to update portfolio")
		return
	}

	utils.RespondJSON(w, http.StatusOK, updatedPortfolio)
}

// DeletePortfolio - DELETE /user/portfolios/delete/{id}
func DeletePortfolio(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value("uid").(string)
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, http.StatusBadRequest, "Portfolio ID is required")
		return
	}

	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase database")
		return
	}

	ref := client.NewRef("portfolios/" + uid + "/" + id)
	err = ref.Delete(context.Background())
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, "Portfolio not found or already deleted")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "Portfolio deleted successfully"})
}
