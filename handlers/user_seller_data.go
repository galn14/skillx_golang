package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"golang-firebase-backend/config"
)

// HandleGetUserAndSellerData fetches user and register seller data
func HandleGetUserAndSellerData(w http.ResponseWriter, r *http.Request) {
	// Get UID from context
	uid := r.Context().Value("uid").(string)

	// Initialize Firebase database client
	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		http.Error(w, `{"error": "Failed to connect to Firebase"}`, http.StatusInternalServerError)
		return
	}

	// Fetch user data
	userRef := client.NewRef("users/" + uid)
	var user map[string]interface{}
	if err := userRef.Get(context.Background(), &user); err != nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	// Fetch registerSeller data
	sellerRef := client.NewRef("registerSellers/" + uid)
	var registerSeller map[string]interface{}
	if err := sellerRef.Get(context.Background(), &registerSeller); err != nil {
		registerSeller = nil // Handle case where no seller data exists
	}

	// Combine data
	response := map[string]interface{}{
		"user":           user,
		"registerSeller": registerSeller,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetUserAndSellerDataByQuery fetches user and register seller data by ID from query parameter
func HandleGetUserAndSellerDataByQuery(w http.ResponseWriter, r *http.Request) {
	// Extract ID from query parameter
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error": "ID is required"}`, http.StatusBadRequest)
		return
	}

	// Initialize Firebase database client
	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		http.Error(w, `{"error": "Failed to connect to Firebase"}`, http.StatusInternalServerError)
		return
	}

	// Fetch user data
	userRef := client.NewRef("users/" + id)
	var user map[string]interface{}
	if err := userRef.Get(context.Background(), &user); err != nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	// Fetch registerSeller data
	sellerRef := client.NewRef("registerSellers/" + id)
	var registerSeller map[string]interface{}
	if err := sellerRef.Get(context.Background(), &registerSeller); err != nil {
		registerSeller = nil // Handle case where no seller data exists
	}

	// Combine data
	response := map[string]interface{}{
		"user":           user,
		"registerSeller": registerSeller,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
