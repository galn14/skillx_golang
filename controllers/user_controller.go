package controllers

import (
	"context"
	"encoding/json"
	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/services"
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

	// Jika major tersedia, ambil titleMajor dari Major collection
	var majorTitle string
	if user.Major != "" {
		var major models.Major
		majorRef := client.NewRef("majors/" + user.Major)
		if err := majorRef.Get(ctx, &major); err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch major data")
			return
		}
		majorTitle = major.TitleMajor
	}

	// Tambahkan major title ke response user
	response := map[string]interface{}{
		"uid":          user.UID,
		"name":         user.Name,
		"email":        user.Email,
		"organization": user.Organization,
		"major": map[string]string{
			"idMajor":    user.Major,
			"titleMajor": majorTitle,
		},
		"language":   user.Language,
		"photo_url":  user.PhotoURL,
		"verified":   user.Verified,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"last_sign_in": func() interface{} {
			if user.LastSignIn.IsZero() {
				return nil
			}
			return user.LastSignIn
		}(),
	}

	utils.RespondJSON(w, http.StatusOK, response)
}
func FetchUserByUID(w http.ResponseWriter, r *http.Request) {
	// Extract UID from the request query parameters
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		utils.RespondError(w, http.StatusBadRequest, "UID is required")
		return
	}

	// Initialize Firebase Realtime Database
	ctx := context.Background()
	dbClient, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Retrieve user data from Firebase
	var user models.User
	userRef := dbClient.NewRef("users/" + uid)
	if err := userRef.Get(ctx, &user); err != nil {
		utils.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Prepare the response data
	response := map[string]interface{}{
		"uid":          user.UID,
		"name":         user.Name,
		"email":        user.Email,
		"organization": user.Organization,
		"major":        user.Major,
		"language":     user.Language,
		"photo_url":    user.PhotoURL,
		"verified":     user.Verified,
		"role":         user.Role,
		"created_at":   user.CreatedAt,
		"last_sign_in": func() interface{} {
			if user.LastSignIn.IsZero() {
				return nil
			}
			return user.LastSignIn
		}(),
	}

	// Send the response back to the frontend
	utils.RespondJSON(w, http.StatusOK, response)
}

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
	var updatedUser map[string]interface{}
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

	// Initialize warnings
	var warnings []string

	// Update fields dynamically based on input
	if name, ok := updatedUser["name"].(string); ok && name != "" {
		existingUser.Name = name
	}
	if organization, ok := updatedUser["organization"].(string); ok && organization != "" {
		existingUser.Organization = organization
	}
	if language, ok := updatedUser["language"].(string); ok && language != "" {
		existingUser.Language = language
	}

	// Dalam bagian Major validation di fungsi UpdateUser
	if major, ok := updatedUser["major"]; ok {
		switch v := major.(type) {
		case string: // major as string (titleMajor)
			if services.IsValidMajorTitle(ctx, v) {
				existingUser.Major = v
			} else {
				log.Printf("Invalid major title: %s, removing major", v)
				existingUser.Major = "" // Remove major if invalid
				warnings = append(warnings, "Provided major is not registered. Major field will not be updated.")
			}
		case map[string]interface{}: // major as object
			if titleMajor, ok := v["titleMajor"].(string); ok && titleMajor != "" {
				if services.IsValidMajorTitle(ctx, titleMajor) {
					existingUser.Major = titleMajor
				} else {
					log.Printf("Invalid major title: %s, removing major", titleMajor)
					existingUser.Major = "" // Remove major if invalid
					warnings = append(warnings, "Provided major is not registered. Major field will not be updated.")
				}
			} else {
				log.Printf("Invalid major format, removing major")
				existingUser.Major = "" // Remove major if invalid format
				warnings = append(warnings, "Invalid major format provided. Major field will not be updated.")
			}
		default:
			log.Printf("Invalid major format, removing major")
			existingUser.Major = "" // Remove major if invalid format
			warnings = append(warnings, "Invalid major format provided. Major field will not be updated.")
		}
	}

	// Write updated user to Firebase
	if err := userRef.Set(ctx, existingUser); err != nil {
		log.Printf("Failed to update user in Firebase: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Failed to update user data")
		return
	}
	log.Println("User updated successfully")

	// Respond with success and warnings (if any)
	response := map[string]interface{}{
		"message":  "User updated successfully",
		"uid":      uid,
		"warnings": warnings,
	}
	utils.RespondJSON(w, http.StatusOK, response)
}

func SearchUsersByName(w http.ResponseWriter, r *http.Request) {
	// Get the search term from query parameters
	searchTerm := r.URL.Query().Get("query")
	if searchTerm == "" {
		utils.RespondError(w, http.StatusBadRequest, "Search term is required")
		return
	}

	ctx := context.Background()

	// Initialize Firebase Realtime Database
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Reference to users node in the database
	usersRef := client.NewRef("users")
	var users map[string]models.User

	// Fetch all users from the database
	if err := usersRef.Get(ctx, &users); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	// Filter users based on the search term
	var matchingUsers []map[string]interface{}
	for id, user := range users {
		if strings.Contains(strings.ToLower(user.Name), strings.ToLower(searchTerm)) {
			matchingUsers = append(matchingUsers, map[string]interface{}{
				"uid":          id,
				"name":         user.Name,
				"email":        user.Email,
				"organization": user.Organization,
				"major":        user.Major,
				"language":     user.Language,
				"photo_url":    user.PhotoURL,
				"verified":     user.Verified,
				"role":         user.Role,
				"created_at":   user.CreatedAt,
				"last_sign_in": user.LastSignIn,
			})
		}
	}

	// Respond with matching users
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"users":   matchingUsers,
	})
}
