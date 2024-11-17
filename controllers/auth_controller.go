package controllers

import (
	"context"
	"encoding/json"
	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"firebase.google.com/go/auth"
)

// RegisterWithEmail handles user registration with bcrypt encryption
func RegisterWithEmail(w http.ResponseWriter, r *http.Request) {
	// Decode body JSON into User model
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate input
	if user.Email == "" || user.Password == "" || user.Name == "" {
		utils.RespondError(w, http.StatusBadRequest, "Name, Email, and Password are required")
		return
	}

	// Hash the password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}
	user.Password = string(hashedPassword)

	// Initialize Firebase Auth
	ctx := context.Background()
	client, err := config.FirebaseApp.Auth(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Auth")
		return
	}

	// Create user in Firebase Authentication
	params := (&auth.UserToCreate{}).
		Email(user.Email).
		DisplayName(user.Name)
	authUser, err := client.CreateUser(ctx, params)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create user: "+err.Error())
		return
	}

	// Save user data to Realtime Database
	dbClient, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	user.UID = authUser.UID
	user.Verified = false // Email is considered not verified
	user.CreatedAt = time.Now()

	ref := dbClient.NewRef("users/" + authUser.UID)
	if err := ref.Set(ctx, user); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to save user data: "+err.Error())
		return
	}

	// Success response
	utils.RespondJSON(w, http.StatusCreated, map[string]string{
		"message": "User registered successfully.",
		"uid":     authUser.UID,
	})
}

// LoginWithEmail handles user login with bcrypt password validation
func LoginWithEmail(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode the request body into the credentials struct
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required fields
	if credentials.Email == "" || credentials.Password == "" {
		utils.RespondError(w, http.StatusBadRequest, "Email and Password are required")
		return
	}

	// Retrieve user from Firebase Realtime Database
	ctx := context.Background()
	dbClient, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Initialize a map to hold the query results
	var users map[string]models.User
	ref := dbClient.NewRef("users")
	if err := ref.OrderByChild("email").EqualTo(credentials.Email).Get(ctx, &users); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to query users: "+err.Error())
		return
	}

	// Check if a user with the given email exists
	if len(users) == 0 {
		utils.RespondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Extract the first user from the query results
	var user models.User
	for _, u := range users {
		user = u
		break
	}

	// Compare the hashed password with the provided password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Successful login response
	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Login successful",
		"uid":     user.UID,
	})
}

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

	// Include Google data and token in the response
	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"user": map[string]interface{}{
			"uid":      user.UID,
			"name":     user.Name,
			"email":    user.Email,
			"photoURL": user.PhotoURL,
		},
		"token": idToken,
	})
}
