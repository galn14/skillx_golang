package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"golang-firebase-backend/config"
)

func HandleAdminVerifySeller(w http.ResponseWriter, r *http.Request) {
	// Decode body request
	var request struct {
		UID    string `json:"uid"`
		Status string `json:"status"` // "accepted" atau "denied"
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validasi input
	if request.UID == "" {
		http.Error(w, "UID is required", http.StatusBadRequest)
		return
	}
	if request.Status != "accepted" && request.Status != "denied" {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	// Inisialisasi Firebase database client
	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		http.Error(w, "Failed to connect to Firebase", http.StatusInternalServerError)
		return
	}

	// Ambil data pengajuan
	ref := client.NewRef("registerSellers/" + request.UID)
	var registerSeller map[string]interface{}
	if err := ref.Get(context.Background(), &registerSeller); err != nil {
		http.Error(w, "RegisterSeller not found", http.StatusNotFound)
		return
	}

	// Update status dan waktu
	registerSeller["status"] = request.Status
	registerSeller["updated_at"] = time.Now().Format(time.RFC3339)
	if request.Status == "accepted" {
		registerSeller["verified"] = true
	}

	// Simpan perubahan
	if err := ref.Set(context.Background(), registerSeller); err != nil {
		http.Error(w, "Failed to update register seller", http.StatusInternalServerError)
		return
	}

	// Jika status "accepted", perbarui user menjadi verified
	if request.Status == "accepted" {
		userRef := client.NewRef("users/" + request.UID)
		if err := userRef.Update(context.Background(), map[string]interface{}{
			"verified": true,
		}); err != nil {
			http.Error(w, "Failed to update user status", http.StatusInternalServerError)
			return
		}
	}

	// Kirim respons sukses
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":         "Seller verification status updated",
		"register_seller": registerSeller,
	})
}
