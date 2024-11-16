package controllers

import (
	"context"
	"encoding/json"
	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"
	"net/http"

	"github.com/google/uuid"
)

// Fetch all skills
func FetchSkills(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Fetch skills from Firebase Realtime Database
	var skills map[string]models.Skill
	ref := client.NewRef("skills")
	if err := ref.Get(ctx, &skills); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch skills")
		return
	}

	// Convert map to slice
	var skillList []models.Skill
	for id, skill := range skills {
		skill.IdSkill = id
		skillList = append(skillList, skill)
	}

	utils.RespondJSON(w, http.StatusOK, skillList)
}

// Show a specific skill
func ShowSkill(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, http.StatusBadRequest, "Skill ID is required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	var skill models.Skill
	ref := client.NewRef("skills/" + id)
	if err := ref.Get(ctx, &skill); err != nil {
		utils.RespondError(w, http.StatusNotFound, "Skill not found")
		return
	}

	skill.IdSkill = id
	utils.RespondJSON(w, http.StatusOK, skill)
}

// Create a new skill
func CreateSkill(w http.ResponseWriter, r *http.Request) {
	var skill models.Skill
	if err := json.NewDecoder(r.Body).Decode(&skill); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if skill.TitleSkills == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "TitleSkills is required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Generate a new unique ID for the skill
	id := uuid.New().String()
	ref := client.NewRef("skills/" + id)
	if err := ref.Set(ctx, &skill); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create skill")
		return
	}

	skill.IdSkill = id
	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    skill,
		"message": "Skill created successfully",
	})
}

// Update an existing skill
func UpdateSkill(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, http.StatusBadRequest, "Skill ID is required")
		return
	}

	var skill models.Skill
	if err := json.NewDecoder(r.Body).Decode(&skill); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if skill.TitleSkills == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "TitleSkills is required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("skills/" + id)
	if err := ref.Update(ctx, map[string]interface{}{
		"TitleSkills": skill.TitleSkills,
	}); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to update skill")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Skill updated successfully",
	})
}

// Delete a skill
func DeleteSkill(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, http.StatusBadRequest, "Skill ID is required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("skills/" + id)
	if err := ref.Delete(ctx); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to delete skill")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Skill deleted successfully",
	})
}
