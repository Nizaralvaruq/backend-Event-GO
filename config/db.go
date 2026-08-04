package config

import (
	"log"
	"os"

	"example.com/event/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DATABASE_URL")

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Printf("DEBUG: Connection error: %v\n", err)
		log.Fatal("Failed to connect to database")
	}

	err = conn.AutoMigrate(&models.Event{}, &models.User{}, &models.Booking{})
	if err != nil {
		log.Fatal("Failed to migrate database")
	}

	DB = conn
	log.Println("Connected to database")
}