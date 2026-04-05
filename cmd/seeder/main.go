package main

import (
	"context"
	"log"
	"os"

	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/database/seeder"
)

func main() {
	// Initialize the application and database connection
	appInstance, err := app.NewApp(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	db := appInstance.DB

	// Seed the database with initial data
	if err := seeder.Run(db.Gorm); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	log.Println("Database seeding completed successfully.")
	os.Exit(0)

}
