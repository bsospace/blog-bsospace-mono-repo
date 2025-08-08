package ai

import (
	"rag-searchbot-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIRepositoryInterface interface {
	CreateChat(chat *models.AIResponse) error
	GetChatByPost(postID string, user *models.User) (*models.AIResponse, error)
	GetChatsByPost(postID string, userID *uuid.UUID, limit, offset int) ([]models.AIResponse, error)
}

type AIRepository struct {
	DB *gorm.DB
}

func NewAIRepository(db *gorm.DB) *AIRepository {
	return &AIRepository{DB: db}
}

func (r *AIRepository) GetChatByPost(postID string, user *models.User) (*models.AIResponse, error) {
	var chat models.AIResponse
	err := r.DB.Preload("User").Preload("Post").
		Where("post_id = ? AND user_id = ?", postID, user.ID).
		First(&chat).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *AIRepository) CreateChat(chat *models.AIResponse) error {
	if err := r.DB.Create(chat).Error; err != nil {
		return err
	}
	return r.DB.Preload("User").Preload("Post").First(chat, "id = ?", chat.ID).Error
}

func (r *AIRepository) GetChatsByPost(postID string, userID *uuid.UUID, limit, offset int) ([]models.AIResponse, error) {
	var chats []models.AIResponse
	db := r.DB.Preload("User").Preload("Post").Where("post_id = ?", postID)
	if userID != nil {
		db = db.Where("user_id = ?", *userID)
	}
	if limit > 0 {
		db = db.Limit(limit)
	}
	if offset > 0 {
		db = db.Offset(offset)
	}
	err := db.Order("used_at asc").Find(&chats).Error
	if err != nil {
		return nil, err
	}
	return chats, nil
}
