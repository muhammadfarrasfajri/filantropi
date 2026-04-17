package config

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var FirebaseAppUser *firebase.App

func InitFirebase() {
	optUser := option.WithCredentialsFile("firebase-key.json")
	appUser, err := firebase.NewApp(context.Background(), nil, optUser)
	if err != nil {
		log.Fatalf("Failed to init Firebase: %v", err)
	}
	FirebaseAppUser = appUser
}
