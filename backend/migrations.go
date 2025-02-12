package main

import (
	"log"

	"github.com/chat-app/database"
	"github.com/chat-app/models"
)

// RunMigrations applies pending database migrations
func RunMigrations() {
	log.Println("running database migrations...")

	err := database.DB.AutoMigrate(
		&models.User{},
		&models.Message{},
	)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database migrations completed successfully!")
}
