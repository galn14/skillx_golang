package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
	"golang-firebase-backend/utils"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Generate conversation key based on user IDs
func generateConversationID(user1, user2 string) string {
	if user1 < user2 {
		return user1 + "_" + user2
	}
	return user2 + "_" + user1
}

// / FetchConversations fetches all conversations for a given user
func FetchConversations(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value("uid")
	if uid == nil {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	// Initialize Firebase database
	ctx := context.Background()
	dbClient, err := config.Database(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error initializing Firebase database: %v", err), http.StatusInternalServerError)
		return
	}

	// Fetch all conversations from Firebase
	ref := dbClient.NewRef("conversations")
	var allConversations map[string]map[string]interface{}
	err = ref.Get(ctx, &allConversations)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching conversations: %v", err), http.StatusInternalServerError)
		return
	}

	// Filter conversations where the user is a participant
	var userConversations []map[string]interface{}
	for conversationID, conversation := range allConversations {
		// Check if participants exist and include the user
		participants, ok := conversation["participants"].([]interface{})
		if !ok {
			continue
		}
		for _, participant := range participants {
			if participant == uid {
				// Ensure "lastMessage" field exists and set default values if missing
				lastMessage, ok := conversation["lastMessage"].(map[string]interface{})
				if !ok {
					lastMessage = map[string]interface{}{
						"messageContent": "",
						"senderID":       "",
						"timestamp":      "",
						"lastMessageId":  "",
					}
					conversation["lastMessage"] = lastMessage
				} else {
					// Set default values for missing fields in "lastMessage"
					if _, exists := lastMessage["messageContent"]; !exists {
						lastMessage["messageContent"] = ""
					}
					if _, exists := lastMessage["senderID"]; !exists {
						lastMessage["senderID"] = ""
					}
					if _, exists := lastMessage["timestamp"]; !exists {
						lastMessage["timestamp"] = ""
					}
					if _, exists := lastMessage["lastMessageId"]; !exists {
						lastMessage["lastMessageId"] = ""
					}
				}

				// Add conversation ID to the conversation data
				conversation["id"] = conversationID
				userConversations = append(userConversations, conversation)
				break
			}
		}
	}

	// Return the filtered conversations
	w.Header().Set("Content-Type", "application/json")
	if len(userConversations) == 0 {
		// Return an empty array if no conversations are found
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    []interface{}{}, // Empty array
		}); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    userConversations,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}

// FetchMessages fetches all messages in a conversation
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

	// Initialize Firebase database
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

// SendMessage sends a new message in a conversation
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

	// Initialize Firebase database
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

	// Update conversation with the latest message
	convRef := client.NewRef("conversations/" + conversationID)
	if err := convRef.Set(ctx, map[string]interface{}{
		"lastMessageId": messageID,
		"lastMessage": map[string]interface{}{
			"senderID":       message.SenderID,
			"messageContent": message.MessageContent,
			"timestamp":      message.Timestamp,
		},
		"participants": []string{message.SenderID, message.ReceiverID},
		"updatedAt":    time.Now(),
	}); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to update conversation")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Message sent successfully",
		"data": map[string]string{
			"id": messageID, // ID unik dari pesan yang baru dikirim
		},
	})

}
func CreateChatRoom(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the context
	uid := r.Context().Value("uid")
	if uid == nil {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized access")
		return
	}

	// Decode the request body to get the other participant's ID
	var payload struct {
		ParticipantID string `json:"participantID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if payload.ParticipantID == "" {
		utils.RespondError(w, http.StatusBadRequest, "Participant ID is required")
		return
	}

	// Generate a unique conversation ID using both user IDs
	currentUserID := uid.(string)
	conversationID := generateConversationID(currentUserID, payload.ParticipantID)

	// Initialize Firebase database
	ctx := context.Background()
	client, err := config.Database(ctx)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to connect to Firebase Database")
		return
	}

	// Check if the conversation already exists
	convRef := client.NewRef("conversations/" + conversationID)
	var existingConversation map[string]interface{}
	if err := convRef.Get(ctx, &existingConversation); err == nil && existingConversation != nil {
		utils.RespondJSON(w, http.StatusConflict, map[string]interface{}{
			"success": false,
			"message": "Chatroom already exists",
			"data": map[string]string{
				"conversationID": conversationID,
			},
		})
		return
	}

	// Create the new chatroom in the database
	timestamp := time.Now().Format(time.RFC3339)
	newChatRoom := map[string]interface{}{
		"participants": []string{currentUserID, payload.ParticipantID},
		"lastMessage": map[string]interface{}{
			"messageContent": "",
			"senderID":       "",
			"timestamp":      "",
		},
		"lastMessageId": "",
		"updatedAt":     timestamp,
	}

	if err := convRef.Set(ctx, newChatRoom); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create chatroom")
		return
	}

	// Respond with the newly created chatroom details
	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Chatroom created successfully",
		"data": map[string]string{
			"conversationID": conversationID,
		},
	})
}
