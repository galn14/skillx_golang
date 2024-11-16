package config

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"

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
