package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/internal/models"
	"time"
)

type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repo: repo}
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

// GetUserProfileOpenId ดึงข้อมูล profile จาก OpenID API
func (s *Service) GetUserProfileOpenId(token string) (*OpenIDProfileData, error) {
	cfg := config.LoadConfig()

	url := fmt.Sprintf("%s/auth/profile", cfg.OpenIDURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var profileResp OpenIDProfileResponse
	if err := json.Unmarshal(body, &profileResp); err != nil {
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

// func (s *Service) UpdateUser(user *models.User) error {
// 	// Update user profile in the database
// 	err := s.Repo.UpdateUser(user)
// 	if err != nil {
// 		return fmt.Errorf("failed to update user profile: %w", err)
// 	}

// 	return nil
// }
