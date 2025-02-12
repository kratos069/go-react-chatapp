package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitializeDatabase creates a GORM database connection
func InitializeDatabase() *gorm.DB {
	// Load environment variables
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	// Create DSN (Database Source Name)
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbSSLMode,
	)

	// Configure GORM with connection pooling and logger
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Enable detailed logs
	})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)                 // Max idle connections
	sqlDB.SetMaxOpenConns(100)                // Max open connections
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Max connection lifetime

	DB = db
	log.Println("Database connection established successfully!")
	return db
}
