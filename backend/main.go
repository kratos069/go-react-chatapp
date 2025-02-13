package main

import (
	"log"
	"os"

	"github.com/chat-app/database"
	"github.com/chat-app/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize database
	database.InitializeDatabase()

	// Run Migrations
	RunMigrations()

	// Create a Fiber app
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		// AllowOrigins:     "http://localhost:5173",  // for development
		AllowOrigins:     os.Getenv("CLIENT_URL"),                       // for Production from environment variables
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",             // Define HTTP methods
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization", // Include additional headers if needed
		AllowCredentials: true,                                          // Enable cookies/credentials sharing
	}))

	// Routes
	RoutesSetup(app, database.DB)

	// WebSocket route
	app.Get("/api/ws", websocket.New(utils.WebSocketHandler))

	// Start server
	serverPort := os.Getenv("SERVER_PORT")
	log.Printf("Server is running on port %s", serverPort)
	log.Fatal(app.Listen(":" + serverPort))
}
