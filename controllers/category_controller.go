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

// Fetch all categories
func FetchCategories(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	var categories map[string]models.Category
	ref := client.NewRef("categories")
	if err := ref.Get(ctx, &categories); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	var categoryList []models.Category
	for id, category := range categories {
		category.IdCategory = id
		categoryList = append(categoryList, category)
	}

	utils.RespondJSON(w, http.StatusOK, categoryList)
}

// Show a specific category by ID
func ShowCategory(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Id string `json:"id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if requestBody.Id == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "Category ID is required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	var category models.Category
	ref := client.NewRef("categories/" + requestBody.Id)
	if err := ref.Get(ctx, &category); err != nil {
		utils.RespondError(w, http.StatusNotFound, "Category not found")
		return
	}

	category.IdCategory = requestBody.Id
	utils.RespondJSON(w, http.StatusOK, category)
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	var categoryInput struct {
		Title      string `json:"title"`
		PhotoUrl   string `json:"photo_url"`
		TitleMajor string `json:"title_major"`
	}

	if err := json.NewDecoder(r.Body).Decode(&categoryInput); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	// Validasi input
	if categoryInput.Title == "" || categoryInput.PhotoUrl == "" || categoryInput.TitleMajor == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "Title, PhotoUrl, and TitleMajor are required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Cari IdMajor berdasarkan TitleMajor
	var majors map[string]models.Major
	refMajors := client.NewRef("majors")
	if err := refMajors.Get(ctx, &majors); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch majors")
		return
	}

	var idMajor string
	for id, major := range majors {
		if major.TitleMajor == categoryInput.TitleMajor {
			idMajor = id
			break
		}
	}

	if idMajor == "" {
		utils.RespondError(w, http.StatusBadRequest, "TitleMajor not found")
		return
	}

	// Buat category
	category := models.Category{
		Title:    categoryInput.Title,
		PhotoUrl: categoryInput.PhotoUrl,
		IdMajor:  idMajor,
	}

	id := uuid.New().String()
	ref := client.NewRef("categories/" + id)
	if err := ref.Set(ctx, &category); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create category")
		return
	}

	category.IdCategory = id
	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    category,
		"message": "Category created successfully",
	})
}

// Delete a category
func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Id string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if requestBody.Id == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "Category ID is required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("categories/" + requestBody.Id)
	if err := ref.Delete(ctx); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to delete category")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Category deleted successfully",
	})
}
func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	idCategory := r.URL.Query().Get("id_category")
	if idCategory == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "Category ID is required")
		return
	}

	var requestBody struct {
		Title    string `json:"title,omitempty"`
		PhotoUrl string `json:"photo_url,omitempty"`
		IdMajor  string `json:"id_major,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if requestBody.Title == "" && requestBody.PhotoUrl == "" && requestBody.IdMajor == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "At least one field (Title, PhotoUrl, IdMajor) must be provided")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("categories/" + idCategory)

	// Fetch existing data
	var existingCategory map[string]interface{}
	if err := ref.Get(ctx, &existingCategory); err != nil {
		utils.RespondError(w, http.StatusNotFound, "Category not found")
		return
	}

	// Hapus field `IdMajor` jika ada
	if _, ok := existingCategory["IdMajor"]; ok {
		if err := ref.Child("IdMajor").Delete(ctx); err != nil {
			fmt.Println("Failed to delete old IdMajor field:", err)
		}
	}

	// Prepare update data
	updateData := make(map[string]interface{})
	if requestBody.Title != "" {
		updateData["title"] = requestBody.Title
	}
	if requestBody.PhotoUrl != "" {
		updateData["photo_url"] = requestBody.PhotoUrl
	}
	if requestBody.IdMajor != "" {
		updateData["id_major"] = requestBody.IdMajor
	}

	// Update Firebase
	if err := ref.Update(ctx, updateData); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to update category")
		return
	}

	// Response JSON
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id_category": idCategory,
			"title":       updateData["title"],
			"photo_url":   updateData["photo_url"],
			"id_major":    updateData["id_major"],
		},
		"message": "Category updated successfully",
	})
}
