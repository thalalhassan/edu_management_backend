package main

import (
	"context"
	"log"
	"os"

	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/database/seed"
)

func main() {
	// Initialize the application and database connection
	appInstance, err := app.NewApp(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	db := appInstance.DB

	command := "" // default command

	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	// Seed the database with initial data
	if err := seed.Seed(db.Gorm, command); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	log.Println("Database seeding completed successfully.")
	os.Exit(0)

}
