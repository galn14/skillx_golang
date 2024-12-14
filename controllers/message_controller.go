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

// Generate unique conversation ID
func generateConversationID(user1, user2 string) string {
	if user1 < user2 {
		return user1 + "_" + user2
	}
	return user2 + "_" + user1
}

// Fetch all conversations for a user
func FetchConversations(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid")
	if uid == nil {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	ctx := context.Background()
	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	var conversations map[string]models.Conversation
	ref := client.NewRef("conversations")
	if err := ref.OrderByChild("participants").EqualTo(uid.(string)).Get(ctx, &conversations); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch conversations")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    conversations,
	})
}

// Fetch messages in a conversation
func FetchMessages(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid")
	if uid == nil {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	conversationID := r.URL.Query().Get("conversationID")
	if conversationID == "" {
		utils.RespondError(w, http.StatusBadRequest, "ConversationID is required")
		return
	}

	ctx := context.Background()
	client, err := config.Database(ctx)
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

	messageList := make([]models.Message, 0, len(messages))
	for id, message := range messages {
		message.ID = id
		messageList = append(messageList, message)
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    messageList,
	})
}

// Send a new message
func SendMessage(w http.ResponseWriter, r *http.Request) {
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

	message.SenderID = uid.(string)
	message.Timestamp = time.Now()
	message.IsRead = false

	conversationID := generateConversationID(message.SenderID, message.ReceiverID)

	ctx := context.Background()
	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	messageID := uuid.New().String()
	ref := client.NewRef("messages/" + conversationID + "/" + messageID)
	if err := ref.Set(ctx, &message); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	// Update conversation
	convRef := client.NewRef("conversations/" + conversationID)
	if err := convRef.Set(ctx, map[string]interface{}{
		"lastMessage":  message,
		"participants": []string{message.SenderID, message.ReceiverID},
		"updatedAt":    time.Now(),
	}); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to update conversation")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Message sent successfully",
	})
}
