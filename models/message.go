package models

import "time"

// Message represents a message entity
type Message struct {
	ID             string    `json:"id"`
	SenderID       string    `json:"senderID"`
	ReceiverID     string    `json:"receiverID"`
	MessageContent string    `json:"messageContent"`
	IsRead         bool      `json:"isRead"`
	Timestamp      time.Time `json:"timestamp"`
}

type Conversation struct {
	ID           string    `json:"id"`
	Participants []string  `json:"participants"`
	LastMessage  Message   `json:"lastMessage"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
