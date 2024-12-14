package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"

	"github.com/google/uuid"
)

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	ctx := context.Background()

	// Get authenticated user's UID from context
	userID := r.Context().Value("uid").(string)

	// Check if the user has the 'seller' role
	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	usersRef := client.NewRef("users/" + userID)
	var user map[string]interface{}

	if err := usersRef.Get(ctx, &user); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch user information")
		return
	}

	if role, ok := user["role"].(string); !ok || role != "seller" {
		utils.RespondError(w, http.StatusForbidden, "Only sellers can create products")
		return
	}

	// Set product fields
	product.UID = uuid.New().String()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	// Save product under the user's node
	ref := client.NewRef("products/" + userID + "/" + product.UID)
	if err := ref.Set(ctx, &product); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create product")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    product,
		"message": "Product created successfully",
	})
}

func FetchProducts(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get authenticated user's UID from context
	userID := r.Context().Value("uid").(string)

	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("products/" + userID)
	var products map[string]models.Product

	if err := ref.Get(ctx, &products); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch products")
		return
	}

	var productList []models.Product
	for id, product := range products {
		product.UID = id
		productList = append(productList, product)
	}

	utils.RespondJSON(w, http.StatusOK, productList)
}

// ViewProduct retrieves a specific product by user name and product name
func ViewProduct(w http.ResponseWriter, r *http.Request) {
	// Get user name and product name from query parameters
	userName := r.URL.Query().Get("name")
	productName := r.URL.Query().Get("product_name")

	// Validate input parameters
	if userName == "" || productName == "" {
		utils.RespondError(w, http.StatusBadRequest, "User name and product name are required")
		return
	}

	ctx := context.Background()

	// Initialize Firebase client
	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Reference the users node to find the user ID by name
	usersRef := client.NewRef("users")
	var users map[string]map[string]interface{}

	// Retrieve all users
	if err := usersRef.Get(ctx, &users); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	// Find the user ID associated with the user name
	var userID string
	for id, user := range users {
		if userNameDB, ok := user["name"].(string); ok && strings.EqualFold(userNameDB, userName) {
			userID = id
			break
		}
	}

	// If no matching user is found, return an error
	if userID == "" {
		utils.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Reference the products node for the specific user
	productsRef := client.NewRef("products/" + userID)
	var products map[string]models.Product

	// Retrieve all products for the user
	if err := productsRef.Get(ctx, &products); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch products")
		return
	}

	// Find the specific product by name
	var product models.Product
	var found bool
	for _, p := range products {
		if strings.EqualFold(p.NameProduct, productName) {
			product = p
			found = true
			break
		}
	}

	// If no matching product is found, return an error
	if !found {
		utils.RespondError(w, http.StatusNotFound, "Product not found")
		return
	}

	// Respond with the product datax
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    product,
	})
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		UID string `json:"uid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if requestBody.UID == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "Product UID is required")
		return
	}

	var updatedProduct map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updatedProduct); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	ctx := context.Background()
	userID := r.Context().Value("uid").(string)

	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("products/" + userID + "/" + requestBody.UID)
	updatedProduct["updated_at"] = time.Now()

	if err := ref.Update(ctx, updatedProduct); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to update product")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Product updated successfully",
	})
}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		UID string `json:"uid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if requestBody.UID == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "Product UID is required")
		return
	}

	ctx := context.Background()
	userID := r.Context().Value("uid").(string)

	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("products/" + userID + "/" + requestBody.UID)
	if err := ref.Delete(ctx); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Product deleted successfully",
	})
}

func SearchProducts(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("query")

	if searchTerm == "" {
		utils.RespondError(w, http.StatusBadRequest, "Search term is required")
		return
	}

	ctx := context.Background()
	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Search users
	usersRef := client.NewRef("users")
	var users map[string]map[string]interface{}

	if err := usersRef.Get(ctx, &users); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	var matchingUsers []string
	for id, user := range users {
		if userName, ok := user["name"].(string); ok && strings.Contains(strings.ToLower(userName), strings.ToLower(searchTerm)) {
			matchingUsers = append(matchingUsers, id)
		}
	}

	// Search products
	productsRef := client.NewRef("products")
	var products map[string]map[string]models.Product

	if err := productsRef.Get(ctx, &products); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch products")
		return
	}

	var filteredProducts []models.Product
	for userID, userProducts := range products {
		// Check if the user ID matches the search query
		if contains(matchingUsers, userID) {
			for _, product := range userProducts {
				filteredProducts = append(filteredProducts, product)
			}
		}

		// Check if the product name matches the search query
		for _, product := range userProducts {
			if strings.Contains(strings.ToLower(product.NameProduct), strings.ToLower(searchTerm)) {
				filteredProducts = append(filteredProducts, product)
			}
		}
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    filteredProducts,
	})
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
