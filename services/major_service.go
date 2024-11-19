package services

import (
	"context"
	"log"

	"golang-firebase-backend/config"
	"golang-firebase-backend/models"
)

// IsValidMajorTitle checks if a given titleMajor exists in the majors collection
func IsValidMajorTitle(ctx context.Context, titleMajor string) bool {
	// Get Database Client
	dbClient, err := config.Database(ctx)
	if err != nil {
		log.Printf("Failed to initialize Firebase Database: %v", err)
		return false
	}

	// Reference the majors collection
	majorsRef := dbClient.NewRef("majors")

	// Fetch all majors
	var majors map[string]models.Major
	if err := majorsRef.Get(ctx, &majors); err != nil {
		log.Printf("Failed to fetch majors: %v", err)
		return false
	}

	// Check if the titleMajor exists
	for _, major := range majors {
		if major.TitleMajor == titleMajor {
			return true
		}
	}

	return false
}
