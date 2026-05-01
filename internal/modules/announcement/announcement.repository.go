package announcement

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

// ─── Repository interface ─────────────────────────────────────────────────────

type Repository interface {
	WithTx(tx *gorm.DB) Repository
	Create(ctx context.Context, a *Announcement) error
	GetByID(ctx context.Context, id uuid.UUID) (*Announcement, error)
	FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*Announcement, int64, error)
	Update(ctx context.Context, a *Announcement) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ─── repositoryImpl ───────────────────────────────────────────────────────────

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) WithTx(tx *gorm.DB) Repository {
	return &repositoryImpl{db: tx}
}

func (r *repositoryImpl) Create(ctx context.Context, a *Announcement) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *repositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*Announcement, error) {
	var a Announcement
	if err := r.db.WithContext(ctx).
		Model(&database.Announcement{}).
		First(&a, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *repositoryImpl) FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*Announcement, int64, error) {
	var announcements []*Announcement
	query := r.db.WithContext(ctx).Model(&database.Announcement{})

	f := q.Filter
	if f.Audience != nil {
		query = query.Where("audience = ?", *f.Audience)
	}
	if f.IsPublished != nil {
		query = query.Where("is_published = ?", *f.IsPublished)
	}
	if f.Search != nil && *f.Search != "" {
		searchTerm := "%" + *f.Search + "%"
		query = query.Where("title ILIKE ?", searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Sort.Apply(query).
		Offset(q.Pagination.Offset).
		Limit(q.Pagination.Limit).
		Find(&announcements).Error
	return announcements, total, err
}

func (r *repositoryImpl) Update(ctx context.Context, a *Announcement) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *repositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&database.Announcement{}, "id = ?", id).Error
}
