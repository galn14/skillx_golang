package services

import (
	"context"
	"errors"
	"log"

	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
)

// GetMajorBySeller retrieves the major associated with a user who is a seller
func GetMajorBySeller(ctx context.Context, userID string) (string, error) {
	// Get Database Client
	dbClient, err := config.Database(ctx)
	if err != nil {
		log.Printf("Failed to initialize Firebase Database: %v", err)
		return "", err
	}

	// Reference the user in the database
	userRef := dbClient.NewRef("users/" + userID)

	// Fetch user data
	var user models.User
	if err := userRef.Get(ctx, &user); err != nil {
		log.Printf("Failed to fetch user for userID %s: %v", userID, err)
		return "", err
	}

	// Check if the user is a seller
	if user.Role != "seller" {
		log.Printf("User %s is not a seller", userID)
		return "", errors.New("user is not a seller")
	}

	// Return the user's major
	if user.Major == "" {
		log.Printf("Seller %s does not have a major assigned", userID)
		return "", errors.New("seller does not have a major assigned")
	}

	return user.Major, nil
}

func GetServiceIDByTitle(ctx context.Context, titleService string) (string, error) {
	// Get Database Client
	dbClient, err := config.Database(ctx)
	if err != nil {
		log.Printf("Failed to initialize Firebase Database: %v", err)
		return "", err
	}

	// Reference the services collection
	servicesRef := dbClient.NewRef("services")

	// Fetch all services
	var services map[string]models.Service
	if err := servicesRef.Get(ctx, &services); err != nil {
		log.Printf("Failed to fetch services: %v", err)
		return "", err
	}

	// Find service by title
	for id, service := range services {
		if service.TitleService == titleService {
			return id, nil
		}
	}

	log.Printf("Service with title '%s' not found", titleService)
	return "", nil
}

// IsValidService checks if a given service exists in the services collection
func IsValidService(ctx context.Context, idService string) bool {
	// Get Database Client
	dbClient, err := config.Database(ctx)
	if err != nil {
		log.Printf("Failed to initialize Firebase Database: %v", err)
		return false
	}

	// Reference the services collection
	serviceRef := dbClient.NewRef("services/" + idService)

	// Fetch the service
	var service models.Service
	if err := serviceRef.Get(ctx, &service); err != nil {
		log.Printf("Service not found: %v", err)
		return false
	}

	return true
}
