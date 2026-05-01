package announcement

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/query_params"
	"gorm.io/gorm"
)

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) WithTx(tx *gorm.DB) Repository {
	args := m.Called(tx)
	return args.Get(0).(Repository)
}

func (m *MockRepository) Create(ctx context.Context, a *Announcement) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id uuid.UUID) (*Announcement, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Announcement), args.Error(1)
}

func (m *MockRepository) FindAll(ctx context.Context, q query_params.Query[FilterParams]) ([]*Announcement, int64, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]*Announcement), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) Update(ctx context.Context, a *Announcement) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

var authorID = uuid.New()

func TestService_Create(t *testing.T) {
	tests := []struct {
		name        string
		req         CreateRequest
		authorID    uuid.UUID
		mockSetup   func(*MockRepository)
		expectError bool
		errorType   interface{}
	}{
		{
			name: "happy path",
			req: CreateRequest{
				Title:    "Test Announcement",
				Body:     "Test Body",
				Audience: database.AnnouncementAudienceAll,
			},
			authorID: authorID,
			mockSetup: func(m *MockRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*announcement.Announcement")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "validation failure - invalid audience",
			req: CreateRequest{
				Title:    "Test",
				Body:     "Body",
				Audience: "INVALID",
			},
			mockSetup:   func(m *MockRepository) {},
			expectError: true,
			errorType:   ValidationError{},
		},
		{
			name: "validation failure - past expires_at",
			req: CreateRequest{
				Title:     "Test",
				Body:      "Body",
				Audience:  database.AnnouncementAudienceAll,
				ExpiresAt: &[]time.Time{time.Now().Add(-1 * time.Hour)}[0],
			},
			mockSetup:   func(m *MockRepository) {},
			expectError: true,
			errorType:   ValidationError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.mockSetup(mockRepo)
			svc := NewService(mockRepo)

			_, err := svc.Create(context.Background(), tt.req, tt.authorID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.IsType(t, tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		userAudiences []AnnouncementAudience
		mockSetup     func(*MockRepository)
		expectError   bool
		errorType     interface{}
	}{
		{
			name:          "happy path",
			id:            "ann-123",
			userAudiences: []AnnouncementAudience{database.AnnouncementAudienceAll},
			mockSetup: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "ann-123").Return(&Announcement{
					Audience: database.AnnouncementAudienceAll,
				}, nil)
			},
			expectError: false,
		},
		{
			name:          "not found",
			id:            "ann-123",
			userAudiences: []AnnouncementAudience{database.AnnouncementAudienceAll},
			mockSetup: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "ann-123").Return((*Announcement)(nil), ErrNotFound)
			},
			expectError: true,
			errorType:   ErrNotFound,
		},
		{
			name:          "unauthorized access",
			id:            "ann-123",
			userAudiences: []AnnouncementAudience{database.AnnouncementAudienceTeachers},
			mockSetup: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "ann-123").Return(&Announcement{
					Audience: database.AnnouncementAudienceStudents,
				}, nil)
			},
			expectError: true,
			errorType:   ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.mockSetup(mockRepo)
			svc := NewService(mockRepo)

			_, err := svc.GetByID(context.Background(), uuid.MustParse(tt.id), tt.userAudiences)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.Equal(t, tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		req         UpdateRequest
		mockSetup   func(*MockRepository)
		expectError bool
		errorType   interface{}
	}{
		{
			name: "happy path",
			id:   "ann-123",
			req: UpdateRequest{
				Title: &[]string{"Updated Title"}[0],
			},
			mockSetup: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "ann-123").Return(&Announcement{
					IsPublished: false,
				}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*announcement.Announcement")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "business rule violation - published",
			id:   "ann-123",
			req: UpdateRequest{
				Title: &[]string{"Updated"}[0],
			},
			mockSetup: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "ann-123").Return(&Announcement{
					IsPublished: true,
				}, nil)
			},
			expectError: true,
			errorType:   BusinessError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.mockSetup(mockRepo)
			svc := NewService(mockRepo)

			_, err := svc.Update(context.Background(), uuid.MustParse(tt.id), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.IsType(t, tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Publish(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		req         PublishRequest
		mockSetup   func(*MockRepository)
		expectError bool
		errorType   interface{}
	}{
		{
			name: "publish draft",
			id:   "ann-123",
			req:  PublishRequest{IsPublished: true},
			mockSetup: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "ann-123").Return(&Announcement{
					IsPublished: false,
				}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*announcement.Announcement")).Return(nil)
			},
			expectError: false,
		},
		{
			name: "business rule violation - already published",
			id:   "ann-123",
			req:  PublishRequest{IsPublished: true},
			mockSetup: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "ann-123").Return(&Announcement{
					IsPublished: true,
				}, nil)
			},
			expectError: true,
			errorType:   BusinessError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.mockSetup(mockRepo)
			svc := NewService(mockRepo)

			_, err := svc.Publish(context.Background(), uuid.MustParse(tt.id), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.IsType(t, tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		mockSetup   func(*MockRepository)
		expectError bool
		errorType   interface{}
	}{
		{
			name: "delete draft",
			id:   "ann-123",
			mockSetup: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "ann-123").Return(&Announcement{
					IsPublished: false,
				}, nil)
				m.On("Delete", mock.Anything, "ann-123").Return(nil)
			},
			expectError: false,
		},
		{
			name: "business rule violation - published",
			id:   "ann-123",
			mockSetup: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "ann-123").Return(&Announcement{
					IsPublished: true,
				}, nil)
			},
			expectError: true,
			errorType:   BusinessError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			tt.mockSetup(mockRepo)
			svc := NewService(mockRepo)

			err := svc.Delete(context.Background(), uuid.MustParse(tt.id))

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.IsType(t, tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

// Test concurrency/duplicate scenario (simplified)
func TestService_Create_Concurrency(t *testing.T) {
	// This is a simplified test; in real scenario, use goroutines and sync
	mockRepo := &MockRepository{}
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*announcement.Announcement")).Return(nil).Once()

	svc := NewService(mockRepo)

	req := CreateRequest{
		Title:    "Concurrent Test",
		Body:     "Body",
		Audience: database.AnnouncementAudienceAll,
	}

	_, err := svc.Create(context.Background(), req, authorID)
	assert.NoError(t, err)

	// Second create should succeed (no unique constraint in this test)
	_, err = svc.Create(context.Background(), req, uuid.New())
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}
