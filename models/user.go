package models

import "time"

// User represents a user entity stored in Firebase
type User struct {
	UID          string    `json:"uid"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Organization string    `json:"organization"`
	Major        string    `json:"major"`
	Language     string    `json:"language"`
	Password     string    `json:"password"`  // Password is not serialized to JSON
	PhotoURL     string    `json:"photo_url"` // Optional
	Verified     bool      `json:"verified"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	LastSignIn   time.Time `json:"last_sign_in,omitempty"`
}

// NewBuyer creates a new User with the buyer role
func NewBuyer(uid, name, email, organization, major, language, password string) *User {
	return &User{
		UID:          uid,
		Name:         name,
		Email:        email,
		Organization: organization,
		Major:        major,
		Language:     language,
		Password:     password,
		Verified:     false,   // Default to not verified
		Role:         "buyer", // Set role as buyer
		CreatedAt:    time.Now(),
		LastSignIn:   time.Time{}, // Empty sign-in time
	}
}
