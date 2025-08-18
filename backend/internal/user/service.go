package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/location"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/social"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ServiceInterface interface {
	RegisterUser(user *models.User) (*models.User, error)
	GetUser(id string) (*models.User, error)
	GetUsers() ([]models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserProfileOpenId(token string) (*OpenIDProfileData, error)
	GetExistingUsername(username string) (bool, error)
	GetUserProfileByUsername(username string, currentUserID *uuid.UUID) (*UserProfileResponse, error)
	UpdateUser(user *models.User) error
	RefreshTokenOpenId(token string) (string, error)
}

type Service struct {
	Repo            RepositoryInterface
	Cache           cache.ServiceInterface
	LocationService *location.LocationService
	SocialService   *social.SocialMediaService
}

func NewService(repo RepositoryInterface, cache cache.ServiceInterface) ServiceInterface {
	return &Service{
		Repo:            repo,
		Cache:           cache,
		LocationService: location.NewLocationService(),
		SocialService:   social.NewSocialMediaService(),
	}
}

// RegisterUser เพิ่ม User ลงในฐานข้อมูล
func (s *Service) RegisterUser(user *models.User) (*models.User, error) {
	err := s.Repo.CreateUser(user)
	return user, err
}

// GetUser ค้นหา User ตาม ID
func (s *Service) GetUser(id string) (*models.User, error) {
	return s.Repo.GetUserByID(id)
}

// GetUsers ดึง Users ทั้งหมด
func (s *Service) GetUsers() ([]models.User, error) {
	return s.Repo.GetUsers()
}

// GetUserByEmail ค้นหา User ตาม Email (ใช้ใน AuthMiddleware)
func (s *Service) GetUserByEmail(email string) (*models.User, error) {
	return s.Repo.GetUserByEmail(email)
}

type OpenIDProfileResponse struct {
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type OpenIDProfileData struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Image     string    `json:"image"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// GetUserProfileOpenId ดึงข้อมูล User Profile จาก OpenID
func (s *Service) GetUserProfileOpenId(token string) (*OpenIDProfileData, error) {
	cfg := config.LoadConfig()
	url := fmt.Sprintf("%s/auth/profile?service=blog", cfg.OpenIDURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-access-token", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var profileResp OpenIDProfileResponse
	if err := json.Unmarshal(responseBody, &profileResp); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := profileResp.Message
		if errMsg == "" {
			errMsg = profileResp.Error
		}
		if errMsg == "" {
			errMsg = "Failed to fetch profile"
		}
		return nil, errors.New(errMsg)
	}

	if profileResp.Data == nil {
		return nil, errors.New("empty profile data")
	}

	// Convert interface{} to map[string]interface{}
	dataMap, ok := profileResp.Data.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid profile data format")
	}

	// Create and populate OpenIDProfileData
	profileData := &OpenIDProfileData{
		ID:       dataMap["id"].(string),
		Username: dataMap["username"].(string),
		Email:    dataMap["email"].(string),
		Image:    dataMap["image"].(string),
	}

	return profileData, nil
}

// get existing username

func (s *Service) GetExistingUsername(username string) (bool, error) {
	result, err := s.Repo.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // Username not found
		}
		return false, fmt.Errorf("failed to check username: %w", err)
	}
	return result, nil
}

func (s *Service) UpdateUser(user *models.User) error {
	// Validate location if provided
	if user.Location != "" {
		if validatedLocation, err := s.LocationService.ValidateLocation(user.Location); err != nil {
			return fmt.Errorf("invalid location: %w", err)
		} else {
			user.Location = validatedLocation
		}
	}

	// Validate social media links
	socialData := map[string]string{
		"github":    user.GitHub,
		"twitter":   user.Twitter,
		"linkedin":  user.LinkedIn,
		"instagram": user.Instagram,
		"facebook":  user.Facebook,
		"youtube":   user.YouTube,
		"discord":   user.Discord,
		"telegram":  user.Telegram,
		"website":   user.Website,
	}

	profiles, validationErrors := s.SocialService.ValidateAllSocialMedia(socialData)
	if len(validationErrors) > 0 {
		return fmt.Errorf("social media validation errors: %s", strings.Join(validationErrors, "; "))
	}

	// Update user with validated social media data
	for platform, profile := range profiles {
		switch platform {
		case "github":
			user.GitHub = profile.Username
		case "twitter":
			user.Twitter = profile.Username
		case "linkedin":
			user.LinkedIn = profile.Username
		case "instagram":
			user.Instagram = profile.Username
		case "facebook":
			user.Facebook = profile.Username
		case "youtube":
			user.YouTube = profile.Username
		case "discord":
			user.Discord = profile.Username
		case "telegram":
			user.Telegram = profile.Username
		case "website":
			user.Website = profile.URL
		}
	}

	err := s.Repo.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	s.Cache.ClearUserCache(user.Email)

	return nil
}

// RefreshTokenOpenId refresh token จาก OpenID
func (s *Service) RefreshTokenOpenId(token string) (string, error) {
	cfg := config.LoadConfig()
	url := fmt.Sprintf("%s/auth/refresh?service=blog", cfg.OpenIDURL)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-refresh-token", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request error: %w", err)
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to refresh token: %s", string(responseBody))
	}

	var response struct {
		Success      bool   `json:"success"`
		Message      string `json:"message"`
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		Error        string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", fmt.Errorf("failed to parse response JSON: %w", err)
	}

	if !response.Success {
		return "", fmt.Errorf("refresh failed: %s", response.Error)
	}

	if response.AccessToken == "" {
		return "", errors.New("access token not found in response")
	}

	return response.AccessToken, nil
}

// GetUserProfileByUsername ดึงข้อมูล User Profile โดย Username
func (s *Service) GetUserProfileByUsername(username string, currentUserID *uuid.UUID) (*UserProfileResponse, error) {
	user, err := s.Repo.GetUserProfileByUsername(username)
	if err != nil {
		return nil, err
	}

	// Check if current user can edit this profile
	canEdit := false
	if currentUserID != nil {
		canEdit = user.ID == *currentUserID
	}

	// Get followers and following count (placeholder for now)
	followers := int64(0)
	following := int64(0)

	profile := &UserProfileResponse{
		Username:  user.UserName,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Avatar:    user.Avatar,
		Bio:       user.Bio,
		Role:      string(user.Role),
		Location:  user.Location,
		Website:   user.Website,
		JoinedAt:  user.CreatedAt.Format("January 2006"),
		Followers: followers,
		Following: following,
		CanEdit:   canEdit,
		SocialMedia: SocialMediaLinks{
			GitHub:    user.GitHub,
			Twitter:   user.Twitter,
			LinkedIn:  user.LinkedIn,
			Instagram: user.Instagram,
			Facebook:  user.Facebook,
			YouTube:   user.YouTube,
			Discord:   user.Discord,
			Telegram:  user.Telegram,
		},
	}

	return profile, nil
}
