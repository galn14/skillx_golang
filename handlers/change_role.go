package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"golang-firebase-backend/config"
)

func HandleChangeRole(w http.ResponseWriter, r *http.Request) {
	// Ambil UID dari context
	uid := r.Context().Value("uid").(string)

	// Decode body request
	var request struct {
		Role string `json:"role"` // "buyer" atau "seller"
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validasi role
	if request.Role != "buyer" && request.Role != "seller" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	// Inisialisasi Firebase database client
	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		http.Error(w, "Failed to connect to Firebase", http.StatusInternalServerError)
		return
	}

	// Ambil data user
	userRef := client.NewRef("users/" + uid)
	var user map[string]interface{}
	if err := userRef.Get(context.Background(), &user); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Jika role adalah "seller", pastikan user sudah verified
	if request.Role == "seller" && (!user["verified"].(bool)) {
		http.Error(w, "User is not verified to become a seller", http.StatusForbidden)
		return
	}

	// Update role
	if err := userRef.Update(context.Background(), map[string]interface{}{
		"role": request.Role,
	}); err != nil {
		http.Error(w, "Failed to update role", http.StatusInternalServerError)
		return
	}

	// Kirim respons sukses
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Role updated successfully",
		"role":    request.Role,
	})
}
