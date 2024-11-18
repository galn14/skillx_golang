package controllers

import (
	"context"
	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"
	"net/http"
	"time"
)

func LoginWithGoogle(w http.ResponseWriter, r *http.Request) {
	idToken := r.Header.Get("Authorization")

	// Extract the token part if prefixed with "Bearer "
	if len(idToken) > 7 && idToken[:7] == "Bearer " {
		idToken = idToken[7:]
	}

	if idToken == "" {
		utils.RespondError(w, http.StatusBadRequest, "ID Token is required")
		return
	}

	ctx := context.Background()

	// Initialize Firebase Auth
	authClient, err := config.FirebaseApp.Auth(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Auth")
		return
	}

	// Verify the ID token
	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "Invalid ID Token")
		return
	}

	uid := token.UID

	// Initialize Firebase Database
	dbClient, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	var user models.User
	userRef := dbClient.NewRef("users/" + uid)

	// Try to fetch the user from the database
	if err := userRef.Get(ctx, &user); err != nil || user.Email == "" {
		// If user data doesn't exist, fetch user data from Firebase Auth
		authUser, err := authClient.GetUser(ctx, uid)
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch user details")
			return
		}

		// Populate the user model with Firebase Auth data
		user = models.User{
			UID:      uid,
			Name:     authUser.DisplayName,
			Email:    authUser.Email,
			PhotoURL: authUser.PhotoURL,
		}

		// Save the new user data in Firebase Database
		if err := userRef.Set(ctx, user); err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to save user data")
			return
		}
	}

	loginTime := time.Now().Format(time.RFC3339)

	// Include Google data, token, and login time in the response
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Login successful",
		"loginTime": loginTime,
		"user": map[string]interface{}{
			"uid":      user.UID,
			"name":     user.Name,
			"email":    user.Email,
			"photoURL": user.PhotoURL,
		},
		"token": idToken,
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	idToken := r.Header.Get("Authorization")

	// Extract the token part if prefixed with "Bearer "
	if len(idToken) > 7 && idToken[:7] == "Bearer " {
		idToken = idToken[7:]
	}

	if idToken == "" {
		utils.RespondError(w, http.StatusBadRequest, "ID Token is required")
		return
	}

	ctx := context.Background()

	// Initialize Firebase Auth
	authClient, err := config.FirebaseApp.Auth(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Auth")
		return
	}

	// Verify the ID token
	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "Invalid ID Token")
		return
	}

	// Revoke the refresh tokens for the user
	if err := authClient.RevokeRefreshTokens(ctx, token.UID); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to logout user")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Logout successful",
	})
}
