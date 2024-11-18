package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"
	"net/http"

	"github.com/google/uuid"
)

// ViewUserPortfolios - GET /user/portfolios/view
func ViewUserPortfolios(w http.ResponseWriter, r *http.Request) {
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

	// Referensi ke portfolio milik user tertentu
	ref := client.NewRef(fmt.Sprintf("portfolios/%s", uid))
	var portfolios map[string]models.Portfolio

	// Ambil seluruh portfolio milik user
	err = ref.Get(context.Background(), &portfolios)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch user portfolios")
		return
	}

	if portfolios == nil {
		portfolios = make(map[string]models.Portfolio)
	}

	// Ubah ke slice untuk respons JSON
	var result []models.Portfolio
	for _, portfolio := range portfolios {
		result = append(result, portfolio)
	}

	utils.RespondJSON(w, http.StatusOK, result)
}

func ViewSpecificUserPortfolios(w http.ResponseWriter, r *http.Request) {
	// Decode request body
	var requestBody struct {
		UserID string `json:"userID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil || requestBody.UserID == "" {
		utils.RespondError(w, http.StatusBadRequest, "UserID is required in the request body")
		return
	}

	userID := requestBody.UserID

	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase database")
		return
	}

	// Referensi ke portfolio user tertentu
	ref := client.NewRef(fmt.Sprintf("portfolios/%s", userID))
	var portfolios map[string]models.Portfolio

	// Ambil seluruh portfolio milik user
	err = ref.Get(context.Background(), &portfolios)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch portfolios for the specified user")
		return
	}

	if portfolios == nil {
		portfolios = make(map[string]models.Portfolio)
	}

	// Ubah ke slice untuk respons JSON
	var result []models.Portfolio
	for _, portfolio := range portfolios {
		result = append(result, portfolio)
	}

	utils.RespondJSON(w, http.StatusOK, result)
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
	if err := ref.Set(context.Background(), portfolio); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create portfolio")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, portfolio)
}

func UpdatePortfolio(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value("uid").(string)
	if !ok || uid == "" {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	var portfolio models.Portfolio
	if err := json.NewDecoder(r.Body).Decode(&portfolio); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase database")
		return
	}

	ref := client.NewRef(fmt.Sprintf("portfolios/%s/%s", portfolio.UserID, portfolio.ID))
	var existingPortfolio models.Portfolio
	if err := ref.Get(context.Background(), &existingPortfolio); err != nil || existingPortfolio.ID == "" {
		utils.RespondError(w, http.StatusNotFound, "Portfolio not found")
		return
	}

	// Logika untuk field kosong
	if portfolio.Title != "" {
		existingPortfolio.Title = portfolio.Title
	}
	if portfolio.Description != "" {
		existingPortfolio.Description = portfolio.Description
	}
	if portfolio.Link != "" {
		existingPortfolio.Link = portfolio.Link
	}
	if portfolio.Photo != "" || portfolio.Photo == "" {
		existingPortfolio.Photo = portfolio.Photo // Tetap kosong jika input kosong
	}
	if portfolio.Type != "" || portfolio.Type == "" {
		existingPortfolio.Type = portfolio.Type // Tetap kosong jika input kosong
	}
	if portfolio.Status != "" || portfolio.Status == "" {
		existingPortfolio.Status = portfolio.Status // Tetap kosong jika input kosong
	}
	if portfolio.DateCreated != "" {
		existingPortfolio.DateCreated = portfolio.DateCreated
	}
	if portfolio.DateEnd != "" || portfolio.DateEnd == "" {
		existingPortfolio.DateEnd = portfolio.DateEnd // Tetap kosong jika input kosong
	}
	existingPortfolio.IsPresent = portfolio.IsPresent

	// Simpan ke Firebase
	if err := ref.Set(context.Background(), existingPortfolio); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to update portfolio")
		return
	}

	utils.RespondJSON(w, http.StatusOK, existingPortfolio)
}

// DeletePortfolio - POST /user/portfolios/delete
func DeletePortfolio(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value("uid").(string)
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	var requestBody struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil || requestBody.ID == "" {
		utils.RespondError(w, http.StatusBadRequest, "Portfolio ID is required in the request body")
		return
	}

	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase database")
		return
	}

	ref := client.NewRef(fmt.Sprintf("portfolios/%s/%s", uid, requestBody.ID))
	if err := ref.Delete(context.Background()); err != nil {
		utils.RespondError(w, http.StatusNotFound, "Portfolio not found or already deleted")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "Portfolio deleted successfully"})
}
