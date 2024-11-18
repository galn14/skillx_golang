package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"golang-firebase-backend/config"
)

func HandleRequestSeller(w http.ResponseWriter, r *http.Request) {
	// Ambil UID dari context
	uid := r.Context().Value("uid").(string)

	// Decode body request
	var request struct {
		Name            string `json:"name"`
		Email           string `json:"email"`
		Organization    string `json:"organization"`
		Major           string `json:"major"`
		PhotoURL        string `json:"photo_url"`
		GraduationMonth string `json:"graduation_month,omitempty"`
		GraduationYear  int    `json:"graduation_year,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Inisialisasi Firebase database client
	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		http.Error(w, "Failed to connect to Firebase", http.StatusInternalServerError)
		return
	}

	// Cek apakah user sudah memiliki pengajuan
	ref := client.NewRef("registerSellers/" + uid)
	var existing map[string]interface{}
	if err := ref.Get(context.Background(), &existing); err == nil && existing != nil {
		http.Error(w, "User has already submitted a request", http.StatusBadRequest)
		return
	}

	// Buat data pengajuan baru
	newRequest := map[string]interface{}{
		"uid":              uid,
		"name":             request.Name,
		"email":            request.Email,
		"organization":     request.Organization,
		"major":            request.Major,
		"photo_url":        request.PhotoURL,
		"status":           "pending",
		"verified":         false,
		"graduation_month": request.GraduationMonth,
		"graduation_year":  request.GraduationYear,
		"created_at":       time.Now().Format(time.RFC3339),
		"updated_at":       time.Now().Format(time.RFC3339),
	}

	// Simpan data ke Firebase
	if err := ref.Set(context.Background(), newRequest); err != nil {
		http.Error(w, "Failed to save request", http.StatusInternalServerError)
		return
	}

	// Kirim respons sukses
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":         "Seller request submitted",
		"register_seller": newRequest,
	})
}