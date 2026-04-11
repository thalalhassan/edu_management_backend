
package seed

import "gorm.io/gorm"

func MustFirst[T any](db *gorm.DB, dest *T) {
	if err := db.First(dest).Error; err != nil {
		panic(err)
	}
}
