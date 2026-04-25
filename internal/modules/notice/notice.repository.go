package notice

import (
	"context"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

// ─── Repository interface ─────────────────────────────────────────────────────

type Repository interface {
	Create(ctx context.Context, n *Notice) error
	GetByID(ctx context.Context, id string) (*Notice, error)
	FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*Notice, int64, error)
	Update(ctx context.Context, n *Notice) error
	Delete(ctx context.Context, id string) error
}

// ─── repositoryImpl ───────────────────────────────────────────────────────────

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) Create(ctx context.Context, n *Notice) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *repositoryImpl) GetByID(ctx context.Context, id string) (*Notice, error) {
	var n Notice
	if err := r.db.WithContext(ctx).
		Model(&database.Notice{}).
		First(&n, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *repositoryImpl) FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*Notice, int64, error) {
	var notices []*Notice
	query := r.db.WithContext(ctx).Model(&database.Notice{})

	f := q.Filter
	if f.Audience != nil {
		query = query.Where("audience = ?", *f.Audience)
	}
	if f.IsPublished != nil {
		query = query.Where("is_published = ?", *f.IsPublished)
	}
	if f.ClassSectionID != nil {
		// NULL means school-wide; a specific ID scopes to that section.
		// Returning both lets a class-section view show both targeted and global notices.
		query = query.Where("class_section_id = ? OR class_section_id IS NULL", *f.ClassSectionID)
	}
	if f.Search != nil {
		like := "%" + *f.Search + "%"
		query = query.Where("title ILIKE ?", like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Sort.Apply(query).
		Offset(q.Pagination.Offset).
		Limit(q.Pagination.Limit).
		Find(&notices).Error
	return notices, total, err
}

func (r *repositoryImpl) Update(ctx context.Context, n *Notice) error {
	return r.db.WithContext(ctx).Where("id = ?", n.ID).Save(n).Error
}

func (r *repositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&database.Notice{}).Error
}
