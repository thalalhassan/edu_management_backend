package user

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/pagination"
	"gorm.io/gorm"
)

type Repository interface {
	// User CRUD
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	FindAllUsers(ctx context.Context, p pagination.Params) ([]*User, int64, error)
	UpdateUser(ctx context.Context, id string, user *User) error
	DeleteUser(ctx context.Context, id string) error

	// Profile creation — each runs inside the same tx as CreateUser
	CreateStudent(ctx context.Context, tx *gorm.DB, student *Student) error
	CreateTeacher(ctx context.Context, tx *gorm.DB, teacher *Teacher) error
	CreateParent(ctx context.Context, tx *gorm.DB, parent *Parent) error
	CreateStaff(ctx context.Context, tx *gorm.DB, staff *Staff) error

	// Transaction helper — exposes the raw *gorm.DB for atomic service calls
	DB() *gorm.DB
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) DB() *gorm.DB { return r.db }

// ──────────────────────────────────────────────────────────────
// User
// ──────────────────────────────────────────────────────────────

func (r *repositoryImpl) CreateUser(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *repositoryImpl) GetUserByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).
		Preload("Teacher").
		Preload("Student").
		Preload("Parent").
		Preload("Staff").
		First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repositoryImpl) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).
		Preload("Teacher").
		Preload("Student").
		Preload("Parent").
		Preload("Staff").
		First(&u, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repositoryImpl) FindAllUsers(ctx context.Context, p pagination.Params) ([]*User, int64, error) {
	var users []*User
	var total int64

	query := r.db.WithContext(ctx).Model(&database.User{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Teacher").
		Preload("Student").
		Preload("Parent").
		Preload("Staff").
		Offset((p.Page - 1) * p.Limit).
		Limit(p.Limit).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *repositoryImpl) UpdateUser(ctx context.Context, id string, user *User) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Save(user).Error
}

func (r *repositoryImpl) DeleteUser(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&database.User{}, "id = ?", id).Error
}

// ──────────────────────────────────────────────────────────────
// Profile creation (all accept an explicit tx so the caller can
// wrap user + profile creation in one transaction)
// ──────────────────────────────────────────────────────────────

func (r *repositoryImpl) CreateStudent(ctx context.Context, tx *gorm.DB, student *Student) error {
	return tx.WithContext(ctx).Create(student).Error
}

func (r *repositoryImpl) CreateTeacher(ctx context.Context, tx *gorm.DB, teacher *Teacher) error {
	return tx.WithContext(ctx).Create(teacher).Error
}

func (r *repositoryImpl) CreateParent(ctx context.Context, tx *gorm.DB, parent *Parent) error {
	return tx.WithContext(ctx).Create(parent).Error
}

func (r *repositoryImpl) CreateStaff(ctx context.Context, tx *gorm.DB, staff *Staff) error {
	return tx.WithContext(ctx).Create(staff).Error
}
