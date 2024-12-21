package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"

	"github.com/google/uuid"
)

// Fungsi untuk mencocokkan string tanpa case sensitivity
func matchStringsIgnoreCase(str1, str2 string) bool {
	return strings.TrimSpace(strings.ToLower(str1)) == strings.TrimSpace(strings.ToLower(str2))
}

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var productInput struct {
		NameProduct string   `json:"nameProduct"`
		Description string   `json:"description"`
		PhotoURL    []string `json:"photo_url"`
		Price       string   `json:"price"`
		IdCategory  string   `json:"idCategory"`
		IdService   string   `json:"idService"`
	}

	if err := json.NewDecoder(r.Body).Decode(&productInput); err != nil {
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

	// Fetch seller data
	sellerRef := client.NewRef("registerSellers/" + userID)
	var seller models.RegisterSeller
	if err := sellerRef.Get(ctx, &seller); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch seller information")
		return
	}

	// Ambil Major dari seller
	majorName := seller.Major
	if majorName == "" {
		utils.RespondError(w, http.StatusForbidden, "Seller must have a valid Major in registerSellers")
		return
	}

	// Fetch Majors
	majorsRef := client.NewRef("majors")
	var majors map[string]models.Major
	if err := majorsRef.Get(ctx, &majors); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch majors")
		return
	}

	var majorID string
	for id, major := range majors {
		if strings.EqualFold(major.TitleMajor, majorName) {
			majorID = id
			break
		}
	}

	if majorID == "" {
		utils.RespondError(w, http.StatusBadRequest, "No matching Major ID found for seller's Major")
		return
	}

	// Fetch Categories
	categoriesRef := client.NewRef("categories/" + productInput.IdCategory)
	var category models.Category
	if err := categoriesRef.Get(ctx, &category); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch category")
		return
	}

	// Validate Category
	if category.IdMajor != majorID {
		fmt.Printf("Category Major Mismatch: Category ID Major: %s, Expected Major ID: %s\n", category.IdMajor, majorID)
		utils.RespondError(w, http.StatusBadRequest, "Selected Category is not part of the seller's Major")
		return
	}

	// Fetch Services
	servicesRef := client.NewRef("services/" + productInput.IdService)
	var service models.Service
	if err := servicesRef.Get(ctx, &service); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch service")
		return
	}

	// Debug service data

	// Validate Service
	if service.IdCategory != productInput.IdCategory {
		fmt.Printf("Service-Category Mismatch: Service ID Category: %s, Provided Category ID: %s\n", service.IdCategory, productInput.IdCategory)
		utils.RespondError(w, http.StatusBadRequest, "Selected Service is not part of the chosen Category")
		return
	}

	// Buat produk
	product := models.Product{
		UID:         uuid.New().String(),
		NameProduct: productInput.NameProduct,
		Description: productInput.Description,
		PhotoURL:    productInput.PhotoURL,
		Price:       productInput.Price,
		Major:       majorName,
		IdCategory:  productInput.IdCategory,
		IdService:   productInput.IdService,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

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
	userID := r.Context().Value("uid").(string)

	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	productsRef := client.NewRef("products/" + userID)
	var products map[string]models.Product
	if err := productsRef.Get(ctx, &products); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch products")
		return
	}

	var productList []models.Product
	for id, product := range products {
		product.UID = id
		productList = append(productList, product)
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    productList,
	})
}

func ViewProduct(w http.ResponseWriter, r *http.Request) {
	// Mengambil query parameter
	userName := r.URL.Query().Get("name")
	productName := r.URL.Query().Get("product_name")

	// Validasi input parameter
	if userName == "" || productName == "" {
		utils.RespondError(w, http.StatusBadRequest, "Both 'name' and 'product_name' query parameters are required")
		return
	}

	ctx := context.Background()
	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Referensi ke registerSellers untuk mendapatkan UID pengguna berdasarkan nama
	sellersRef := client.NewRef("registerSellers")
	var sellers map[string]models.RegisterSeller

	if err := sellersRef.Get(ctx, &sellers); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch sellers")
		return
	}

	// Mencari UID pengguna berdasarkan nama
	var userID string
	for id, seller := range sellers {
		if strings.EqualFold(seller.Name, userName) {
			userID = id
			break
		}
	}

	// Jika pengguna tidak ditemukan
	if userID == "" {
		utils.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Referensi ke produk berdasarkan UID pengguna
	productsRef := client.NewRef("products/" + userID)
	var products map[string]models.Product

	if err := productsRef.Get(ctx, &products); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch products")
		return
	}

	// Mencari produk berdasarkan nama
	var product models.Product
	var found bool
	for _, p := range products {
		if strings.EqualFold(p.NameProduct, productName) {
			product = p
			found = true
			break
		}
	}

	// Jika produk tidak ditemukan
	if !found {
		utils.RespondError(w, http.StatusNotFound, "Product not found")
		return
	}

	// Mengembalikan respons produk
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    product,
	})
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	// Ambil UID dari query parameter
	productUID := r.URL.Query().Get("uid")

	// Validasi UID
	if productUID == "" {
		utils.RespondError(w, http.StatusBadRequest, "Product UID is required in query parameter")
		return
	}

	// Ambil data yang ingin diperbarui dari body request
	var updateInput struct {
		Description string   `json:"description,omitempty"`
		PhotoURL    []string `json:"photo_url,omitempty"`
		Price       string   `json:"price,omitempty"`
		IdCategory  string   `json:"idCategory,omitempty"`
		IdService   string   `json:"idService,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateInput); err != nil {
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

	// Ambil produk yang ada berdasarkan UID
	productRef := client.NewRef("products/" + userID + "/" + productUID)
	var existingProduct models.Product
	if err := productRef.Get(ctx, &existingProduct); err != nil {
		utils.RespondError(w, http.StatusNotFound, "Product not found")
		return
	}

	// Validasi dan update hanya field yang diperbolehkan
	updates := map[string]interface{}{}
	if updateInput.Description != "" {
		updates["description"] = updateInput.Description
	}
	if updateInput.PhotoURL != nil {
		updates["photo_url"] = updateInput.PhotoURL
	}
	if updateInput.Price != "" {
		updates["price"] = updateInput.Price
	}
	if updateInput.IdCategory != "" {
		// Validasi kategori baru
		categoryRef := client.NewRef("categories/" + updateInput.IdCategory)
		var category models.Category
		if err := categoryRef.Get(ctx, &category); err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid category ID")
			return
		}
		updates["idCategory"] = updateInput.IdCategory
	}
	if updateInput.IdService != "" {
		// Validasi layanan baru
		serviceRef := client.NewRef("services/" + updateInput.IdService)
		var service models.Service
		if err := serviceRef.Get(ctx, &service); err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid service ID")
			return
		}
		if service.IdCategory != updateInput.IdCategory {
			utils.RespondError(w, http.StatusBadRequest, "Service does not belong to the selected category")
			return
		}
		updates["idService"] = updateInput.IdService
	}

	updates["updated_at"] = time.Now()

	// Terapkan pembaruan
	if err := productRef.Update(ctx, updates); err != nil {
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

	productRef := client.NewRef("products/" + userID + "/" + requestBody.UID)
	if err := productRef.Delete(ctx); err != nil {
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
