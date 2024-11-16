package controllers

import (
	"encoding/json"
	"net/http"

	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"

	"golang.org/x/net/context"
)

func AddUserSkill(w http.ResponseWriter, r *http.Request) {
	// Retrieve the UID from context
	uid := r.Context().Value("uid")
	if uid == nil {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	// Parse the request body for skill data
	var userSkill models.UserSkill
	if err := json.NewDecoder(r.Body).Decode(&userSkill); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Ensure the UserId from the request matches the UID in context
	if userSkill.UserId != uid {
		utils.RespondError(w, http.StatusUnauthorized, "User ID does not match authentication token")
		return
	}

	// Process the skill addition for the user (UID)
	// Example logic: Save skill to Firebase Realtime Database or Firestore
	// Save the skill to the Firebase Realtime Database (or Firestore, as needed)
	ctx := context.Background()
	dbClient, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Save the user skill data
	ref := dbClient.NewRef("user_skills/" + userSkill.UserId + "/" + userSkill.IdSkill)
	if err := ref.Set(ctx, userSkill); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to save user skill data")
		return
	}

	// Respond with success
	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Skill added successfully",
	})
}
