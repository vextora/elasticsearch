package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDB() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		getEnv("POSTGRES_HOST", ""),
		getEnv("POSTGRES_USER", ""),
		getEnv("POSTGRES_PASSWORD", ""),
		getEnv("POSTGRES_DB", ""),
		getEnv("POSTGRES_PORT", ""),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	log.Println("Connected to database")
	return db
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
