package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"golang-firebase-backend/config"
)

// HandleGetAllSellers fetches all the registerSeller data
func HandleGetAllSellers(w http.ResponseWriter, r *http.Request) {
	// Initialize Firebase database client
	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		http.Error(w, `{"error": "Failed to connect to Firebase"}`, http.StatusInternalServerError)
		return
	}

	// Fetch all registerSeller data
	sellerRef := client.NewRef("registerSellers")
	var sellers map[string]map[string]interface{}
	if err := sellerRef.Get(context.Background(), &sellers); err != nil {
		http.Error(w, `{"error": "Failed to fetch sellers"}`, http.StatusInternalServerError)
		return
	}

	// Prepare the response data (convert the map into a slice)
	var sellerList []map[string]interface{}
	for _, seller := range sellers {
		sellerList = append(sellerList, seller)
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sellerList)
}
