package seed

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/pkg/appcrypto"
	"gorm.io/gorm"
)

// SeedUsers seeds one user per role plus realistic profiles for
// teacher, student, parent, and staff.
// Safe to re-run — skips rows that already exist by email.
func SeedUsers(db *gorm.DB) error {
	log.Println("[seeder] seeding users...")

	password, err := appcrypto.BcryptHash("Salsabeel")
	if err != nil {
		return fmt.Errorf("seeder.SeedUsers.Hash: %w", err)
	}

	// ── Profiles ──────────────────────────────────────────────

	employees := []database.Employee{
		{
			EmployeeCode:   "EMP001",
			FirstName:      "Rajesh",
			LastName:       "Kumar",
			Gender:         database.GenderMale,
			Category:       database.EmployeeCategoryTeacher,
			Designation:    "Mathematics Teacher",
			Phone:          ptr("9876543210"),
			Qualification:  ptr("M.Sc Mathematics"),
			Specialization: ptr("Mathematics"),
			JoiningDate:    date("2018-06-01"),
			IsActive:       true,
		},
		{
			EmployeeCode:   "EMP002",
			FirstName:      "Priya",
			LastName:       "Nair",
			Gender:         database.GenderFemale,
			Category:       database.EmployeeCategoryTeacher,
			Designation:    "English Teacher",
			Phone:          ptr("9876543211"),
			Qualification:  ptr("M.A English"),
			Specialization: ptr("English Literature"),
			JoiningDate:    date("2019-07-15"),
			IsActive:       true,
		},
	}

	students := []database.Student{
		{
			AdmissionNo:   "ADM2024001",
			FirstName:     "Arjun",
			LastName:      "Sharma",
			DOB:           date("2010-04-12"),
			Gender:        database.GenderMale,
			Status:        database.StudentStatusActive,
			Phone:         ptr("9123456789"),
			AdmissionDate: date("2024-06-01"),
		},
		{
			AdmissionNo:   "ADM2024002",
			FirstName:     "Meera",
			LastName:      "Pillai",
			DOB:           date("2011-08-22"),
			Gender:        database.GenderFemale,
			Status:        database.StudentStatusActive,
			Phone:         ptr("9123456790"),
			AdmissionDate: date("2024-06-01"),
		},
	}

	parents := []database.Parent{
		{
			FirstName:    "Suresh",
			LastName:     "Sharma",
			Relationship: "Father",
			Phone:        "9000000001",
			Occupation:   ptr("Engineer"),
		},
		{
			FirstName:    "Latha",
			LastName:     "Pillai",
			Relationship: "Mother",
			Phone:        "9000000002",
			Occupation:   ptr("Teacher"),
		},
	}

	staff := []database.Employee{
		{
			EmployeeCode: "STF001",
			FirstName:    "Mohan",
			LastName:     "Das",
			Gender:       database.GenderMale,
			Category:     database.EmployeeCategoryStaff,
			Designation:  "Accountant",
			Phone:        ptr("9000000010"),
			JoiningDate:  date("2020-01-10"),
			IsActive:     true,
		},
	}

	// ── Users (role → profile mapping) ───────────────────────

	type userSeed struct {
		email   string
		role    database.SystemRole
		profile any // *Employee | *Student | *Parent | nil
	}

	seeds := []userSeed{
		{email: "superadmin@school.com", role: database.SystemRoleSuperAdmin, profile: nil},
		{email: "admin@school.com", role: database.SystemRoleAdmin, profile: nil},
		{email: "principal@school.com", role: database.SystemRolePrincipal, profile: nil},
		{email: "rajesh.kumar@school.com", role: database.SystemRoleTeacher, profile: &employees[0]},
		{email: "priya.nair@school.com", role: database.SystemRoleTeacher, profile: &employees[1]},
		{email: "arjun.sharma@school.com", role: database.SystemRoleStudent, profile: &students[0]},
		{email: "meera.pillai@school.com", role: database.SystemRoleStudent, profile: &students[1]},
		{email: "suresh.sharma@school.com", role: database.SystemRoleParent, profile: &parents[0]},
		{email: "latha.pillai@school.com", role: database.SystemRoleParent, profile: &parents[1]},
		{email: "mohan.das@school.com", role: database.SystemRoleStaff, profile: &staff[0]},
	}

	for _, s := range seeds {
		s := s

		// Skip if user already exists
		var count int64
		db.Model(&database.User{}).Where("email = ?", s.email).Count(&count)
		if count > 0 {
			log.Printf("[seeder] skipping %s — already exists", s.email)
			continue
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			// Create or get role
			roleSlug := string(s.role)
			role := &database.Role{
				Slug:     roleSlug,
				Name:     strings.Title(strings.ReplaceAll(roleSlug, "_", " ")),
				IsSystem: true,
			}
			if err := tx.FirstOrCreate(role, database.Role{Slug: roleSlug}).Error; err != nil {
				return err
			}

			u := &database.User{
				Email:        s.email,
				PasswordHash: password,
				RoleID:       role.ID,
				IsActive:     true,
			}

			switch p := s.profile.(type) {
			case *database.Employee:
				if err := tx.Create(p).Error; err != nil {
					return fmt.Errorf("create employee (%s): %w", s.email, err)
				}
				u.EmployeeID = &p.ID

			case *database.Student:
				if err := tx.Create(p).Error; err != nil {
					return fmt.Errorf("create student (%s): %w", s.email, err)
				}
				u.StudentID = &p.ID

			case *database.Parent:
				p.Email = &s.email
				if err := tx.Create(p).Error; err != nil {
					return fmt.Errorf("create parent (%s): %w", s.email, err)
				}
				u.ParentID = &p.ID
			}

			if err := tx.Create(u).Error; err != nil {
				return fmt.Errorf("create user (%s): %w", s.email, err)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("seeder.SeedUsers transaction: %w", err)
		}

		log.Printf("[seeder] created %-12s → %s", s.role, s.email)
	}

	log.Println("[seeder] users done")
	return nil
}

// ── helpers ───────────────────────────────────────────────────

func ptr(s string) *string { return &s }

func date(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(fmt.Sprintf("seeder.date: invalid date %q: %v", s, err))
	}
	return t
}
