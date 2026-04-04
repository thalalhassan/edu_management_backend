package main

import (
	"context"
	"log"
	"os"

	"github.com/thalalhassan/edu_management/internal/app"
	"github.com/thalalhassan/edu_management/internal/database"
)

func main() {
	// Initialize the application and database connection
	appInstance, err := app.NewApp(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	db := appInstance.DB

	command := os.Args[1]

	switch command {
	case "up":
		migration_up(db)

	case "down":
		migration_down(db)
	default:
		log.Fatalf("!!! Unknown command: %s", command)
	}

}

func migration_up(db *database.DB) {
	// Perform database migration
	if err := database.MigrateUp(db.Gorm); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed successfully")

}

func migration_down(db *database.DB) {

	if err := database.MigrateDown(db.Gorm); err != nil {
		log.Fatalf("Failed to rollback database: %v", err)
	}

	log.Println("Database rollback completed successfully")
}
