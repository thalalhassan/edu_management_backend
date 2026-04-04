package database

import (
	"gorm.io/gorm"
)

func MigrateUp(DB *gorm.DB) error {
	return DB.AutoMigrate(
		AllModels()...,
	)
}

func MigrateDown(DB *gorm.DB) error {
	for _, table := range AllModels() {
		if err := DB.Migrator().DropTable(&table); err != nil {
			return err
		}
	}

	return nil

}
