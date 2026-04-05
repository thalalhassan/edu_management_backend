package seeder

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// Run executes all seeders in dependency order.
// Call this from your main.go or a dedicated seed command.
func Run(db *gorm.DB) error {
	log.Println("[seeder] starting...")

	seeders := []struct {
		name string
		fn   func(*gorm.DB) error
	}{
		{"users", SeedUsers},
		// add more here as you build them:
		// {"departments", SeedDepartments},
		// {"standards",   SeedStandards},
		// {"subjects",    SeedSubjects},
	}

	for _, s := range seeders {
		log.Printf("[seeder] running: %s", s.name)
		if err := s.fn(db); err != nil {
			return fmt.Errorf("seeder.Run(%s): %w", s.name, err)
		}
	}

	log.Println("[seeder] all done")
	return nil
}
