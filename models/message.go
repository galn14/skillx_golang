package models

import "time"

// Message represents a message entity
type Message struct {
	ID             string    `json:"id"`
	SenderID       string    `json:"sender_id"`
	ReceiverID     string    `json:"receiver_id"`
	MessageTitle   string    `json:"message_title"`
	MessageContent string    `json:"message_content"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Sender represents the user sending the message (for relationship purposes)
type Sender struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Receiver represents the user receiving the message (for relationship purposes)
type Receiver struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
