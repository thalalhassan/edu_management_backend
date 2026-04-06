package leave

import (
	"context"
	"time"

	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

// ─── Repository interface ─────────────────────────────────────────────────────

type Repository interface {
	Create(ctx context.Context, l *TeacherLeave) error
	GetByID(ctx context.Context, id string) (*TeacherLeave, error)
	FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*TeacherLeave, int64, error)
	Update(ctx context.Context, l *TeacherLeave) error
	Delete(ctx context.Context, id string) error

	// HasOverlap checks whether the teacher already has a non-withdrawn leave
	// request whose date range overlaps with [fromDate, toDate].
	// excludeID is used during updates to skip the record being edited.
	HasOverlap(ctx context.Context, teacherID string, fromDate, toDate time.Time, excludeID string) (bool, error)
}

// ─── repositoryImpl ───────────────────────────────────────────────────────────

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) Create(ctx context.Context, l *TeacherLeave) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *repositoryImpl) GetByID(ctx context.Context, id string) (*TeacherLeave, error) {
	var l TeacherLeave
	if err := r.db.WithContext(ctx).
		Model(&database.TeacherLeave{}).
		First(&l, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *repositoryImpl) FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*TeacherLeave, int64, error) {
	var leaves []*TeacherLeave
	query := r.db.WithContext(ctx).Model(&database.TeacherLeave{})

	f := q.Filter
	if f.TeacherID != nil {
		query = query.Where("teacher_id = ?", *f.TeacherID)
	}
	if f.Status != nil {
		query = query.Where("status = ?", *f.Status)
	}
	// date_from / date_to filter against the leave's own from_date / to_date range —
	// returns any leave that overlaps with the requested window.
	if f.DateFrom != nil {
		query = query.Where("to_date >= ?", f.DateFrom.Format("2006-01-02"))
	}
	if f.DateTo != nil {
		query = query.Where("from_date <= ?", f.DateTo.Format("2006-01-02"))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Sort.Apply(query).
		Offset(q.Pagination.Offset).
		Limit(q.Pagination.Limit).
		Find(&leaves).Error
	return leaves, total, err
}

func (r *repositoryImpl) Update(ctx context.Context, l *TeacherLeave) error {
	return r.db.WithContext(ctx).Save(l).Error
}

func (r *repositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&database.TeacherLeave{}, "id = ?", id).Error
}

// HasOverlap uses the half-open interval check:
//
//	existing.from_date < requested.to_date AND existing.to_date > requested.from_date
//
// Only non-WITHDRAWN leaves are considered — a withdrawn leave frees the dates.
func (r *repositoryImpl) HasOverlap(ctx context.Context, teacherID string, fromDate, toDate time.Time, excludeID string) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&database.TeacherLeave{}).
		Where("teacher_id = ?", teacherID).
		Where("status != ?", database.LeaveStatusRejected).
		Where("from_date < ? AND to_date > ?",
			toDate.Format("2006-01-02"),
			fromDate.AddDate(0, 0, -1).Format("2006-01-02"), // shift by -1 day so boundary dates are inclusive
		)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
