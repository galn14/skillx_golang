package controllers

import (
	"context"
	"encoding/json"
	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"
	"log"
	"net/http"
	"strings"
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

// UpdateUser handles updating user data in Firebase Realtime Database
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request on %s", r.URL.Path)
	ctx := context.Background()

	// Extract Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		log.Printf("Invalid or missing Authorization header")
		http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
		return
	}
	idToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Verify ID token
	client, err := config.FirebaseApp.Auth(ctx)
	if err != nil {
		log.Printf("Failed to initialize Firebase Auth: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		log.Printf("Invalid or expired token: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	uid := token.Subject
	log.Printf("Verified UID: %s", uid)

	// Decode the request payload
	var updatedUser models.User
	if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Initialize Firebase Database
	dbClient, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Reference the user in the database
	userRef := dbClient.NewRef("users/" + uid)
	log.Printf("Updating Firebase user: %s", uid)

	// Validate existing user
	var existingUser models.User
	if err := userRef.Get(ctx, &existingUser); err != nil {
		log.Printf("User not found: %v", err)
		utils.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Update fields
	existingUser.Name = updatedUser.Name
	existingUser.Email = updatedUser.Email
	existingUser.Organization = updatedUser.Organization
	existingUser.Major = updatedUser.Major
	existingUser.Language = updatedUser.Language

	// Write updated user to Firebase
	if err := userRef.Set(ctx, existingUser); err != nil {
		log.Printf("Failed to update user in Firebase: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to update user data")
		return
	}
	log.Println("User updated successfully")

	// Respond with success
	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "User updated successfully",
		"uid":     uid,
	})
}
