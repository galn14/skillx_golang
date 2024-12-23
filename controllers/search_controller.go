package controllers

import (
	"context"
	"net/http"
	"strings"

	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"

	"firebase.google.com/go/db"
)

// SearchController handles searching for users and products
// SearchController handles searching for users and products
func SearchController(w http.ResponseWriter, r *http.Request) {
	// Get the search term from query parameters
	searchTerm := r.URL.Query().Get("query")
	if searchTerm == "" {
		utils.RespondError(w, http.StatusBadRequest, "Search term is required")
		return
	}

	searchTerm = strings.ToLower(searchTerm)
	ctx := context.Background()

	// Initialize Firebase Database
	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Fetch matching users
	users, userErr := searchUsers(ctx, client, searchTerm)
	if userErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to search users")
		return
	}

	// Fetch matching products
	products, productErr := searchProducts(ctx, client, searchTerm, users)
	if productErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to search products")
		return
	}

	// Respond with results
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"users":    users,
		"products": products,
	})
}

// searchUsers finds users matching the search term
func searchUsers(ctx context.Context, client *db.Client, searchTerm string) ([]map[string]interface{}, error) {
	usersRef := client.NewRef("users")
	var users map[string]models.User

	if err := usersRef.Get(ctx, &users); err != nil {
		return nil, err
	}

	var matchingUsers []map[string]interface{}
	for id, user := range users {
		if strings.Contains(strings.ToLower(user.Name), searchTerm) {
			matchingUsers = append(matchingUsers, map[string]interface{}{
				"id":           id, // Include the user ID here
				"name":         user.Name,
				"email":        user.Email,
				"organization": user.Organization,
				"major":        user.Major,
				"language":     user.Language,
				"photo_url":    user.PhotoURL,
				"verified":     user.Verified,
				"role":         user.Role,
				"created_at":   user.CreatedAt,
				"last_sign_in": user.LastSignIn,
			})
		}
	}

	return matchingUsers, nil
}

// searchProducts finds products matching the search term or owned by matching users
func searchProducts(ctx context.Context, client *db.Client, searchTerm string, matchingUsers []map[string]interface{}) ([]models.Product, error) {
	productsRef := client.NewRef("products")
	var products map[string]map[string]models.Product

	if err := productsRef.Get(ctx, &products); err != nil {
		return nil, err
	}

	userIDs := make(map[string]bool)
	for _, user := range matchingUsers {
		if uid, ok := user["id"].(string); ok {
			userIDs[uid] = true
		}
	}

	var filteredProducts []models.Product
	for userID, userProducts := range products {
		// Include products owned by matching users
		if userIDs[userID] {
			for _, product := range userProducts {
				filteredProducts = append(filteredProducts, product)
			}
		}

		// Include products with names matching the search term
		for _, product := range userProducts {
			if strings.Contains(strings.ToLower(product.NameProduct), searchTerm) {
				filteredProducts = append(filteredProducts, product)
			}
		}
	}

	return filteredProducts, nil
}
