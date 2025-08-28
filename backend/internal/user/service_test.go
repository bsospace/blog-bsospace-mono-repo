package user

import (
	"context"
	"errors"
	"rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/internal/models"
	"mime/multipart"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(id string) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUsers() ([]models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByUsername(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetUserProfileByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// Mock Cache Service
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Delete(key string) {
	m.Called(key)
}

func (m *MockCacheService) ClearUserCache(email string) {
	m.Called(email)
}

func (m *MockCacheService) SetUserCache(email string, user interface{}) error {
	args := m.Called(email, user)
	return args.Error(0)
}

func (m *MockCacheService) GetUserCache(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockCacheService) GetString(ctx context.Context, key string) (string, bool) {
	args := m.Called(ctx, key)
	return args.String(0), args.Bool(1)
}

func (m *MockCacheService) Get(ctx context.Context, key string) (interface{}, bool) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Bool(1)
}

func (m *MockCacheService) Clear() {
	m.Called()
}

func (m *MockCacheService) SetWarpKey(email string, warpKey string) error {
	args := m.Called(email, warpKey)
	return args.Error(0)
}

func (m *MockCacheService) GetWarpKey(email string) (string, bool) {
	args := m.Called(email)
	return args.String(0), args.Bool(1)
}

func (m *MockCacheService) ClearWarpKey(email string) {
	m.Called(email)
}

// Mock Media Service
type MockMediaService struct {
	mock.Mock
}

func (m *MockMediaService) CreateMedia(fileHeader *multipart.FileHeader, user *models.User, postID *uuid.UUID) (*models.ImageUpload, error) {
	args := m.Called(fileHeader, user, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ImageUpload), args.Error(1)
}

func (m *MockMediaService) DeleteFromChibisafe(image *models.ImageUpload) error {
	args := m.Called(image)
	return args.Error(0)
}

func (m *MockMediaService) GetImagesByPostID(postID uuid.UUID) ([]models.ImageUpload, error) {
	args := m.Called(postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ImageUpload), args.Error(1)
}

func (m *MockMediaService) UpdateImageUsage(image *models.ImageUpload) error {
	args := m.Called(image)
	return args.Error(0)
}

func (m *MockMediaService) DeleteUnusedImages() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMediaService) GetImageByURL(imageURL string) (*models.ImageUpload, error) {
	args := m.Called(imageURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ImageUpload), args.Error(1)
}

func (m *MockMediaService) UploadToChibisafe(fileHeader *multipart.FileHeader) (media.ChibisafeResponse, error) {
	args := m.Called(fileHeader)
	if args.Get(0) == nil {
		return media.ChibisafeResponse{}, args.Error(1)
	}
	return args.Get(0).(media.ChibisafeResponse), args.Error(1)
}

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name          string
		user          *models.User
		mockBehavior  func(*MockUserRepository)
		expectedUser  *models.User
		expectedError error
	}{
		{
			name: "successful registration",
			user: &models.User{
				Email:    "test@example.com",
				UserName: "testuser",
				Role:     models.NormalUser,
			},
			mockBehavior: func(repo *MockUserRepository) {
				repo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedUser: &models.User{
				Email:    "test@example.com",
				UserName: "testuser",
				Role:     models.NormalUser,
			},
			expectedError: nil,
		},
		{
			name: "repository error",
			user: &models.User{
				Email:    "test@example.com",
				UserName: "testuser",
			},
			mockBehavior: func(repo *MockUserRepository) {
				repo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(errors.New("database error"))
			},
			expectedUser:  &models.User{Email: "test@example.com", UserName: "testuser"},
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockCache := &MockCacheService{}
			mockMedia := &MockMediaService{}

			tt.mockBehavior(mockRepo)

			service := NewService(mockRepo, mockCache, mockMedia)
			result, err := service.RegisterUser(tt.user)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser.Email, result.Email)
				assert.Equal(t, tt.expectedUser.UserName, result.UserName)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockBehavior  func(*MockUserRepository)
		expectedUser  *models.User
		expectedError error
	}{
		{
			name:   "successful get user",
			userID: "123e4567-e89b-12d3-a456-426614174000",
			mockBehavior: func(repo *MockUserRepository) {
				expectedUser := &models.User{
					ID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Email:    "test@example.com",
					UserName: "testuser",
				}
				repo.On("GetUserByID", "123e4567-e89b-12d3-a456-426614174000").Return(expectedUser, nil)
			},
			expectedUser: &models.User{
				ID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Email:    "test@example.com",
				UserName: "testuser",
			},
			expectedError: nil,
		},
		{
			name:   "user not found",
			userID: "123e4567-e89b-12d3-a456-426614174000",
			mockBehavior: func(repo *MockUserRepository) {
				repo.On("GetUserByID", "123e4567-e89b-12d3-a456-426614174000").Return(nil, errors.New("user not found"))
			},
			expectedUser:  nil,
			expectedError: errors.New("user not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockCache := &MockCacheService{}
			mockMedia := &MockMediaService{}

			tt.mockBehavior(mockRepo)

			service := NewService(mockRepo, mockCache, mockMedia)
			result, err := service.GetUser(tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser.ID, result.ID)
				assert.Equal(t, tt.expectedUser.Email, result.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		mockBehavior  func(*MockUserRepository)
		expectedUser  *models.User
		expectedError error
	}{
		{
			name:  "successful get user by email",
			email: "test@example.com",
			mockBehavior: func(repo *MockUserRepository) {
				expectedUser := &models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}
				repo.On("GetUserByEmail", "test@example.com").Return(expectedUser, nil)
			},
			expectedUser: &models.User{
				Email: "test@example.com",
			},
			expectedError: nil,
		},
		{
			name:  "user not found by email",
			email: "nonexistent@example.com",
			mockBehavior: func(repo *MockUserRepository) {
				repo.On("GetUserByEmail", "nonexistent@example.com").Return(nil, errors.New("user not found"))
			},
			expectedUser:  nil,
			expectedError: errors.New("user not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockCache := &MockCacheService{}
			mockMedia := &MockMediaService{}

			tt.mockBehavior(mockRepo)

			service := NewService(mockRepo, mockCache, mockMedia)
			result, err := service.GetUserByEmail(tt.email)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser.Email, result.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetUsers(t *testing.T) {
	tests := []struct {
		name          string
		mockBehavior  func(*MockUserRepository)
		expectedUsers []models.User
		expectedError error
	}{
		{
			name: "successful get users",
			mockBehavior: func(repo *MockUserRepository) {
				expectedUsers := []models.User{
					{ID: uuid.New(), Email: "user1@example.com", UserName: "user1"},
					{ID: uuid.New(), Email: "user2@example.com", UserName: "user2"},
				}
				repo.On("GetUsers").Return(expectedUsers, nil)
			},
			expectedUsers: []models.User{
				{Email: "user1@example.com", UserName: "user1"},
				{Email: "user2@example.com", UserName: "user2"},
			},
			expectedError: nil,
		},
		{
			name: "repository error",
			mockBehavior: func(repo *MockUserRepository) {
				repo.On("GetUsers").Return(nil, errors.New("database error"))
			},
			expectedUsers: nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockCache := &MockCacheService{}
			mockMedia := &MockMediaService{}

			tt.mockBehavior(mockRepo)

			service := NewService(mockRepo, mockCache, mockMedia)
			result, err := service.GetUsers()

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, len(tt.expectedUsers))
				for i, user := range result {
					assert.Equal(t, tt.expectedUsers[i].Email, user.Email)
					assert.Equal(t, tt.expectedUsers[i].UserName, user.UserName)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
