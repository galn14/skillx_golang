package config

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"

	"google.golang.org/api/option"
)

var FirebaseApp *firebase.App

func InitializeFirebaseApp() (*firebase.App, error) {
	config := &firebase.Config{
		DatabaseURL: "https://skillx-butterscoth.firebaseio.com/",
	}

	opt := option.WithCredentialsFile("skillx-butterscoth-firebase-adminsdk-eivk1-06ffcd28f7.json")
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	FirebaseApp = app

	return app, err
}

// Database initializes and returns the Firebase Realtime Database client
func Database(ctx context.Context) (*db.Client, error) {
	if FirebaseApp == nil {
		_, err := InitializeFirebaseApp()
		if err != nil {
			return nil, fmt.Errorf("error initializing Firebase App: %v", err)
		}
	}

	client, err := FirebaseApp.Database(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase Database: %v", err)
	}

	return client, nil
}
