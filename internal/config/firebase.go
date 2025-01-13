package config

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var Client *auth.Client

func InitializeFirebase() {
	// Initialize Firebase app
	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile("serviceAccountKey.json"))
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v", err)
	}

	// Get Auth client from Firebase App
	Client, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Firebase Auth client: %v", err)
	}
}
