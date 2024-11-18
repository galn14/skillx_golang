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

// Fetch all messages for the current user
func FetchMessages(w http.ResponseWriter, r *http.Request) {
	// Retrieve the UID from context
	uid := r.Context().Value("uid")
	if uid == nil {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Fetch messages for the user (either as sender or receiver)
	var messages map[string]models.Message
	ref := client.NewRef("messages")
	if err := ref.OrderByChild("ReceiverID").EqualTo(uid).Get(ctx, &messages); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch messages")
		return
	}

	// Convert map to slice
	var messageList []models.Message
	for id, message := range messages {
		message.ID = id
		messageList = append(messageList, message)
	}

	utils.RespondJSON(w, http.StatusOK, messageList)
}

// Show a specific message for the current user
func ShowMessage(w http.ResponseWriter, r *http.Request) {
	// Retrieve the UID from context
	uid := r.Context().Value("uid")
	if uid == nil {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, http.StatusBadRequest, "Message ID is required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Get message from Firebase
	var message models.Message
	ref := client.NewRef("messages/" + id)
	if err := ref.Get(ctx, &message); err != nil {
		utils.RespondError(w, http.StatusNotFound, "Message not found")
		return
	}

	// Ensure the message belongs to the user
	if message.SenderID != uid && message.ReceiverID != uid {
		utils.RespondError(w, http.StatusUnauthorized, "You are not authorized to view this message")
		return
	}

	message.ID = id
	utils.RespondJSON(w, http.StatusOK, message)
}

// Create a new message
func CreateMessage(w http.ResponseWriter, r *http.Request) {
	// Retrieve the UID from context
	uid := r.Context().Value("uid")
	if uid == nil {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	var message models.Message
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	// Validation
	if message.ReceiverID == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "ReceiverID is required")
		return
	}

	// Ensure the sender matches the authenticated user
	if message.SenderID != "" && message.SenderID != uid {
		utils.RespondError(w, http.StatusUnauthorized, "SenderID does not match authentication token")
		return
	}

	// If SenderID is not provided, assign the authenticated user as the sender
	if message.SenderID == "" {
		message.SenderID = uid.(string)
	}

	// Save the message to Firebase
	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Generate a new unique ID for the message
	id := uuid.New().String()
	ref := client.NewRef("messages/" + id)
	if err := ref.Set(ctx, &message); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create message")
		return
	}

	message.ID = id
	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    message,
		"message": "Message created successfully",
	})
}

// Update an existing message
func UpdateMessage(w http.ResponseWriter, r *http.Request) {
	// Retrieve the UID from context
	uid := r.Context().Value("uid")
	if uid == nil {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, http.StatusBadRequest, "Message ID is required")
		return
	}

	var message models.Message
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	// Validation
	if message.MessageContent == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "MessageContent is required")
		return
	}

	// Ensure the message belongs to the authenticated user
	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("messages/" + id)
	if err := ref.Get(ctx, &message); err != nil {
		utils.RespondError(w, http.StatusNotFound, "Message not found")
		return
	}

	// Ensure that the user is either the sender or receiver of the message
	if message.SenderID != uid && message.ReceiverID != uid {
		utils.RespondError(w, http.StatusUnauthorized, "You are not authorized to update this message")
		return
	}

	// Update message content
	if err := ref.Update(ctx, map[string]interface{}{
		"MessageContent": message.MessageContent,
		"MessageTitle":   message.MessageTitle,
	}); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to update message")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Message updated successfully",
	})
}

// Delete a message
func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	// Retrieve the UID from context
	uid := r.Context().Value("uid")
	if uid == nil {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondError(w, http.StatusBadRequest, "Message ID is required")
		return
	}

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Fetch the message first to ensure it's valid for the user
	var message models.Message
	ref := client.NewRef("messages/" + id)
	if err := ref.Get(ctx, &message); err != nil {
		utils.RespondError(w, http.StatusNotFound, "Message not found")
		return
	}

	// Ensure the user is either the sender or receiver of the message
	if message.SenderID != uid && message.ReceiverID != uid {
		utils.RespondError(w, http.StatusUnauthorized, "You are not authorized to delete this message")
		return
	}

	// Delete the message
	if err := ref.Delete(ctx); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to delete message")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Message deleted successfully",
	})
}
