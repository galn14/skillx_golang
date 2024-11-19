package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/services"
	"golang-firebase-backend/utils"

	"github.com/google/uuid"
)

// SearchProducts searches products based on query parameters
// SearchProducts allows searching products using a JSON body
func SearchProducts(w http.ResponseWriter, r *http.Request) {
	// Retrieve the search term from query parameters
	searchTerm := r.URL.Query().Get("title") // "title" is the search parameter

	// If searchTerm is empty, return an error
	if searchTerm == "" {
		utils.RespondError(w, http.StatusBadRequest, "Search term is required")
		return
	}

	// Split the search term into individual words (by spaces)
	searchWords := strings.Fields(searchTerm)

	// Initialize Firebase client
	ctx := context.Background()
	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Reference the products node in Firebase
	ref := client.NewRef("products")
	var products map[string]models.Product

	// Retrieve all products from Firebase
	if err := ref.Get(ctx, &products); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch products")
		return
	}

	// Filter products based on the presence of each search word in NameProduct (case-insensitive)
	var filteredProducts []models.Product
	for _, product := range products {
		// Initialize a flag to check if all words match
		matches := true

		// Check if each word in the search query exists in the product name
		for _, word := range searchWords {
			if !strings.Contains(strings.ToLower(product.NameProduct), strings.ToLower(word)) {
				matches = false
				break
			}
		}

		// If the product matches all the words in the search query, add it to the result
		if matches {
			filteredProducts = append(filteredProducts, product)
		}
	}

	// If no products match, return an empty list or a message
	if len(filteredProducts) == 0 {
		utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"message": "No products found matching the search terms",
			"data":    filteredProducts,
		})
		return
	}

	// Respond with filtered products
	utils.RespondJSON(w, http.StatusOK, filteredProducts)
}

// FetchProducts retrieves all products
func FetchProducts(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("products")
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

// ViewProduct retrieves a specific product by ID
func ViewProduct(w http.ResponseWriter, r *http.Request) {
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
	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("products/" + requestBody.UID)
	var product models.Product

	if err := ref.Get(ctx, &product); err != nil {
		utils.RespondError(w, http.StatusNotFound, "Product not found")
		return
	}

	product.UID = requestBody.UID
	utils.RespondJSON(w, http.StatusOK, product)
}

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product

	// Logging raw body data for debugging
	log.Println("Decoding JSON input...")

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		log.Printf("Failed to decode input: %v", err)
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	log.Printf("Decoded Product Input: %+v", product)

	ctx := context.Background()

	// Get user ID from context
	userID := r.Context().Value("uid").(string)
	log.Printf("User ID: %s", userID)

	// Fetch idMajor from seller role
	idMajor, err := services.GetMajorBySeller(ctx, userID)
	if err != nil {
		log.Printf("Error fetching major for seller: %v", err)
		utils.RespondError(w, http.StatusBadRequest, "User is not a seller or does not have a valid major")
		return
	}
	product.IdMajor = idMajor
	log.Printf("ID Major: %s", idMajor)

	// Handle service validation
	if product.IdService == "" && product.TitleService != "" {
		idService, err := services.GetServiceIDByTitle(ctx, product.TitleService)
		if err != nil || idService == "" {
			log.Printf("Invalid titleService: %v", err)
			utils.RespondError(w, http.StatusBadRequest, "Invalid or unregistered titleService")
			return
		}
		product.IdService = idService
		log.Printf("ID Service: %s", idService)
	}

	// Set timestamps
	product.UID = uuid.New().String()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	// Save product to Firebase
	client, err := config.Database(ctx)
	if err != nil {
		log.Printf("Failed to connect to Firebase Database: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("products/" + product.UID)
	if err := ref.Set(ctx, &product); err != nil {
		log.Printf("Failed to save product to Firebase: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create product")
		return
	}

	log.Println("Product created successfully")

	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    product,
		"message": "Product created successfully",
	})
}

// UpdateProduct updates an existing product
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
	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("products/" + requestBody.UID)

	// Update timestamp
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

// DeleteProduct deletes a product by ID
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
	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("products/" + requestBody.UID)
	if err := ref.Delete(ctx); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Product deleted successfully",
	})
}
