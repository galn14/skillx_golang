package controllers

import (
	"context"
	"encoding/json"
	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Helper function to generate a conversation ID based on two user IDs
func generateConversationID(user1, user2 string) string {
	// Ensure the conversation ID is always sorted (to prevent mismatched ordering)
	if user1 < user2 {
		return "chatroom_" + user1 + "_" + user2
	}
	return "chatroom_" + user2 + "_" + user1
}

// Fetch all messages for the current user in a specific chatroom
// Fetch all messages for a specific conversation
func FetchMessages(w http.ResponseWriter, r *http.Request) {
	// Retrieve the UID from context (authenticated user)
	uid := r.Context().Value("uid")
	if uid == nil {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	partnerID := r.URL.Query().Get("partnerID")
	if partnerID == "" {
		utils.RespondError(w, http.StatusBadRequest, "PartnerID is required")
		return
	}

	// Generate conversationID based on user IDs (sender and receiver)
	senderID := uid.(string)
	conversationID := generateConversationID(senderID, partnerID)

	// Fetch messages for the user and the conversation
	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	var messages map[string]models.Message
	ref := client.NewRef("messages/" + conversationID)
	if err := ref.Get(ctx, &messages); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch messages")
		return
	}

	// Convert map to slice of messages
	var messageList []models.Message
	for id, message := range messages {
		message.ID = id
		messageList = append(messageList, message)
	}

	// Respond with the list of messages
	utils.RespondJSON(w, http.StatusOK, messageList)
}

// Show a specific message in a chatroom
func ShowMessage(w http.ResponseWriter, r *http.Request) {
	// Retrieve the UID from context
	uid := r.Context().Value("uid")
	if uid == nil {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	// Get the other user's ID and message ID
	partnerID := r.URL.Query().Get("partnerID")
	messageID := r.URL.Query().Get("messageID")
	if partnerID == "" || messageID == "" {
		utils.RespondError(w, http.StatusBadRequest, "PartnerID and MessageID are required")
		return
	}

	// Generate the conversation ID
	conversationID := generateConversationID(uid.(string), partnerID)

	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Get message from Firebase
	var message models.Message
	ref := client.NewRef("messages/" + conversationID + "/" + messageID)
	if err := ref.Get(ctx, &message); err != nil {
		utils.RespondError(w, http.StatusNotFound, "Message not found")
		return
	}

	// Ensure the message belongs to the user
	if message.SenderID != uid && message.ReceiverID != uid {
		utils.RespondError(w, http.StatusUnauthorized, "You are not authorized to view this message")
		return
	}

	message.ID = messageID
	utils.RespondJSON(w, http.StatusOK, message)
}

// Create a new message in a chatroom
// Create a new message
func CreateMessage(w http.ResponseWriter, r *http.Request) {
	// Retrieve the UID from context (authenticated user)
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

	// Ensure ReceiverID is provided
	if message.ReceiverID == "" {
		utils.RespondError(w, http.StatusUnprocessableEntity, "ReceiverID is required")
		return
	}

	// Ensure SenderID matches the authenticated user
	if message.SenderID != "" && message.SenderID != uid {
		utils.RespondError(w, http.StatusUnauthorized, "SenderID does not match authentication token")
		return
	}

	// Assign the SenderID to the authenticated user if not provided
	if message.SenderID == "" {
		message.SenderID = uid.(string)
	}

	// Generate conversationID based on SenderID and ReceiverID
	senderID := message.SenderID
	receiverID := message.ReceiverID
	conversationID := generateConversationID(senderID, receiverID)

	// Set the creation time and update time
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	// Save the message in Firebase
	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Generate a new unique ID for the message
	messageID := uuid.New().String()
	ref := client.NewRef("messages/" + conversationID + "/" + messageID)

	if err := ref.Set(ctx, &message); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create message")
		return
	}

	message.ID = messageID

	// Respond with the message details
	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    message,
		"message": "Message created successfully",
	})
}

// Update an existing message in a chatroom
func UpdateMessage(w http.ResponseWriter, r *http.Request) {
	// Retrieve the UID from context
	uid := r.Context().Value("uid")
	if uid == nil {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	// Get the other user's ID and message ID
	partnerID := r.URL.Query().Get("partnerID")
	messageID := r.URL.Query().Get("messageID")
	if partnerID == "" || messageID == "" {
		utils.RespondError(w, http.StatusBadRequest, "PartnerID and MessageID are required")
		return
	}

	// Parse the incoming message content for update
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

	// Generate the conversation ID
	conversationID := generateConversationID(uid.(string), partnerID)

	// Ensure the message exists in the Firebase DB
	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("messages/" + conversationID + "/" + messageID)
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

// Delete a message in a chatroom
func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	// Retrieve the UID from context
	uid := r.Context().Value("uid")
	if uid == nil {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	// Get the other user's ID and message ID
	partnerID := r.URL.Query().Get("partnerID")
	messageID := r.URL.Query().Get("messageID")
	if partnerID == "" || messageID == "" {
		utils.RespondError(w, http.StatusBadRequest, "PartnerID and MessageID are required")
		return
	}

	// Generate the conversation ID
	conversationID := generateConversationID(uid.(string), partnerID)

	// Delete the message
	ctx := context.Background()
	client, err := config.FirebaseApp.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	ref := client.NewRef("messages/" + conversationID + "/" + messageID)
	if err := ref.Delete(ctx); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to delete message")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Message deleted successfully",
	})
}
