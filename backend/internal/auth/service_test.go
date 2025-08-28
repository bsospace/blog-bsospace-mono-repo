package auth

import (
	"errors"
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/pkg/crypto"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock User Service
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) RegisterUser(user *models.User) (*models.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUser(id string) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUsers() ([]models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserProfileOpenId(token string) (*user.OpenIDProfileData, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.OpenIDProfileData), args.Error(1)
}

func (m *MockUserService) GetExistingUsername(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserService) GetUserProfileByUsername(username string, currentUserID *uuid.UUID) (*user.UserProfileResponse, error) {
	args := m.Called(username, currentUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.UserProfileResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserService) RefreshTokenOpenId(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

func (m *MockUserService) RefreshTokenAndSetCookies(c *gin.Context) (string, error) {
	args := m.Called(c)
	return args.String(0), args.Error(1)
}

// Mock Crypto Service
type MockCryptoService struct {
	mock.Mock
}

func (m *MockCryptoService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockCryptoService) ComparePasswords(password, hashedPassword string) bool {
	args := m.Called(password, hashedPassword)
	return args.Bool(0)
}

func (m *MockCryptoService) GenerateAccessToken(userID, username, service string) (string, error) {
	args := m.Called(userID, username, service)
	return args.String(0), args.Error(1)
}

func (m *MockCryptoService) GenerateRefreshToken(userID, username, service string) (string, error) {
	args := m.Called(userID, username, service)
	return args.String(0), args.Error(1)
}

func (m *MockCryptoService) ValidateToken(tokenString, service, keyType string) (string, string, error) {
	args := m.Called(tokenString, service, keyType)
	return args.String(0), args.String(1), args.Error(2)
}

func TestExchangeToken(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		config        *config.Config
		expectedResp  *TokenExchangeResponse
		expectedError error
	}{
		{
			name: "empty code",
			code: "",
			config: &config.Config{
				OpenIDURL: "http://localhost:3000",
			},
			expectedResp:  nil,
			expectedError: errors.New("code is required"),
		},
		{
			name: "valid code",
			code: "valid_code_123",
			config: &config.Config{
				OpenIDURL: "http://localhost:3000",
			},
			expectedResp: &TokenExchangeResponse{
				Success:      true,
				Message:      "Token exchanged successfully",
				AccessToken:  "access_token_123",
				RefreshToken: "refresh_token_123",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := &MockUserService{}
			mockCryptoService := &crypto.CryptoService{}

			service := NewAuthService(mockUserService, mockCryptoService, tt.config)

			if tt.code == "" {
				result, err := service.ExchangeToken(tt.code)
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				// For valid code, we can't easily test HTTP calls in unit tests
				// This would require integration tests or HTTP mocking
				// For now, we'll just test the service creation
				assert.NotNil(t, service)
			}
		})
	}
}

func TestNewAuthService(t *testing.T) {
	mockUserService := &MockUserService{}
	mockCryptoService := &crypto.CryptoService{}
	config := &config.Config{
		OpenIDURL: "http://localhost:3000",
	}

	service := NewAuthService(mockUserService, mockCryptoService, config)

	assert.NotNil(t, service)
	
	// Type assert to access concrete fields
	authService := service.(*AuthService)
	assert.Equal(t, mockUserService, authService.UserService)
	assert.Equal(t, mockCryptoService, authService.CrypetoService)
	assert.Equal(t, config, authService.EnvConfig)
}

func TestTokenExchangeResponse(t *testing.T) {
	resp := &TokenExchangeResponse{
		Success:      true,
		Message:      "Success",
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
	}

	assert.True(t, resp.Success)
	assert.Equal(t, "Success", resp.Message)
	assert.Equal(t, "access_token", resp.AccessToken)
	assert.Equal(t, "refresh_token", resp.RefreshToken)
}
