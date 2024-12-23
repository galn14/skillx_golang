package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"golang-firebase-backend/config"
)

// HandleUpdateAboutMe updates the "about_me" field for a specific seller
func HandleUpdateAboutMe(w http.ResponseWriter, r *http.Request) {
	// Extract Bearer token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"error": "Authorization header missing"}`, http.StatusUnauthorized)
		return
	}

	// Validate Bearer token format
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		http.Error(w, `{"error": "Invalid Authorization header format"}`, http.StatusUnauthorized)
		return
	}
	idToken := tokenParts[1]

	// Initialize Firebase Auth client
	authClient, err := config.FirebaseApp.Auth(context.Background())
	if err != nil {
		http.Error(w, `{"error": "Failed to initialize Firebase Auth"}`, http.StatusInternalServerError)
		return
	}

	// Verify the ID token
	token, err := authClient.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		http.Error(w, `{"error": "Invalid or expired token"}`, http.StatusUnauthorized)
		return
	}

	// Extract UID from the token
	uid := token.UID

	// Parse the request body to get the new "about_me"
	var reqBody struct {
		AboutMe string `json:"about_me"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate input
	if reqBody.AboutMe == "" {
		http.Error(w, `{"error": "AboutMe cannot be empty"}`, http.StatusBadRequest)
		return
	}

	// Initialize Firebase database client
	client, err := config.FirebaseApp.Database(context.Background())
	if err != nil {
		http.Error(w, `{"error": "Failed to connect to Firebase"}`, http.StatusInternalServerError)
		return
	}

	// Reference the specific seller in Firebase
	sellerRef := client.NewRef("registerSellers").Child(uid)

	// Update the "about_me" field
	updateData := map[string]interface{}{
		"about_me": reqBody.AboutMe,
	}
	if err := sellerRef.Update(context.Background(), updateData); err != nil {
		http.Error(w, `{"error": "Failed to update AboutMe"}`, http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "AboutMe updated successfully"}`))
}
