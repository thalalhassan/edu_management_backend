package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/pkg/appcrypto"
	"gorm.io/gorm"
)

type ResetAndBootstrapSeeder struct{}

func (s *ResetAndBootstrapSeeder) Name() string {
	return "reset_and_bootstrap"
}

func (s *ResetAndBootstrapSeeder) Run(ctx context.Context, db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {

		// 1. Disable FK checks (Postgres safe via CASCADE)
		tables := []string{
			"user_refresh_token",
			"role_permission",
			"permission",

			"student_parent",
			"student_enrollment",
			"attendance",
			"teacher_attendance",

			"teacher_assignment",
			"time_table",

			"exam_result",
			"exam_schedule",
			"exam",

			"fee_record",
			"fee_structure",

			// "salary_record",
			// "salary_structure",

			"notice",

			"class_section",
			"standard",
			"department",
			"subject",
			"academic_year",

			"student",
			"parent",
			"teacher",
			"staff",

			"user",
		}

		for _, table := range tables {
			if err := tx.Exec(fmt.Sprintf(`TRUNCATE TABLE "%s" RESTART IDENTITY CASCADE`, table)).Error; err != nil {
				return err
			}
		}

		passwordHash, err := appcrypto.BcryptHash("Salsabeel")
		if err != nil {
			return fmt.Errorf("seeder.SeedUsers.Hash: %w", err)
		}

		// Create super admin role first
		superAdminRole := &database.Role{
			Slug:        "super_admin",
			Name:        "Super Admin",
			Description: "Full system access",
			IsSystem:    true,
			Priority:    100,
		}
		if err := tx.FirstOrCreate(superAdminRole, database.Role{Slug: "super_admin"}).Error; err != nil {
			return err
		}

		superAdmin := &database.User{
			Email:        "superadmin@school.com",
			PasswordHash: passwordHash,
			RoleID:       superAdminRole.ID,
			IsActive:     true,
		}

		if err := tx.Create(superAdmin).Error; err != nil {
			return err
		}

		// // 3. Seed minimal permissions (optional but recommended)
		// permissions := []Permission{
		// 	{Resource: "all", Action: PermissionActionManage, Description: "Full access"},
		// }

		// for _, p := range permissions {
		// 	if err := tx.Create(&p).Error; err != nil {
		// 		return err
		// 	}

		// 	rp := RolePermission{
		// 		UserID:       superAdmin.ID,
		// 		PermissionID: p.ID,
		// 	}

		// 	if err := tx.Create(&rp).Error; err != nil {
		// 		return err
		// 	}
		// }

		// 4. Minimal Academic Year (system requires it)
		year := database.AcademicYear{
			Name:      "2025-2026",
			StartDate: time.Now(),
			EndDate:   time.Now().AddDate(1, 0, 0),
			IsActive:  true,
		}

		if err := tx.Create(&year).Error; err != nil {
			return err
		}

		return nil
	})
}
