package services

import (
	"chat-app/internal/repository"
	"fmt"
	"log"
	"regexp"
)

func ValidateUsername(username string) error {
	re := regexp.MustCompile(`^[a-zA-Z0-9\-_.]{3,10}$`)
	if !re.MatchString(username) {
		return fmt.Errorf("Error: Username must be 3-10 characters long" +
			" and contain only letters,\ndigits, '-', '_', or '.'.")
	}
	return nil
}

func RegisterUserIfNotExists(username string) {
	exists := repository.CheckUserExists(username)
	if !exists {
		log.Printf("creating user: %s", username)
		repository.CreateUser(username)
	}
}

func RegisterConnection(username string) {
	log.Printf("user %s connected", username)
	repository.LogConnection(username, "connected")
}

func RegisterDisconnection(username string) {
	log.Printf("user %s disconected", username)
	repository.LogConnection(username, "disconnected")
}

func HandleMessage(username, message string) {
	log.Printf("[%s]: %s", username, message)
	repository.SaveMessage(username, message)
}
