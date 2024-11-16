package controllers

import (
	"context"
	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"
	"net/http"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	// Ambil UID dari query parameter
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		utils.RespondError(w, http.StatusBadRequest, "UID is required")
		return
	}

	// Inisialisasi Firebase Realtime Database
	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Ambil data user dari Realtime Database
	var user models.User
	ref := client.NewRef("users/" + uid)
	if err := ref.Get(ctx, &user); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch user data")
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}
