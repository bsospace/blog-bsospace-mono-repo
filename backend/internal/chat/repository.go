package chat

import (
	"rag-searchbot-backend/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RepositoryInterface interface {
	CreateMessage(msg *models.AIResponse) error
	GetMessagesBetweenUsers(user1, user2 uuid.UUID, limit, offset int) ([]models.AIResponse, error)
	GetMessagesByRoom(roomID uuid.UUID, limit, offset int) ([]models.AIResponse, error)
	MarkAsRead(messageID uuid.UUID, readAt time.Time) error
}

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) RepositoryInterface {
	return &Repository{DB: db}
}

func (r *Repository) CreateMessage(msg *models.AIResponse) error {
	return r.DB.Create(msg).Error
}

func (r *Repository) GetMessagesBetweenUsers(user1, user2 uuid.UUID, limit, offset int) ([]models.AIResponse, error) {
	var messages []models.AIResponse
	err := r.DB.Where(
		"(user_id = ? AND post_id = ?) OR (user_id = ? AND post_id = ?)",
		user1, user2, user2, user1,
	).Order("used_at desc").Limit(limit).Offset(offset).Find(&messages).Error
	return messages, err
}

func (r *Repository) GetMessagesByRoom(roomID uuid.UUID, limit, offset int) ([]models.AIResponse, error) {
	var messages []models.AIResponse
	err := r.DB.Where("post_id = ?", roomID).Order("used_at desc").Limit(limit).Offset(offset).Find(&messages).Error
	return messages, err
}

func (r *Repository) MarkAsRead(messageID uuid.UUID, readAt time.Time) error {
	return r.DB.Model(&models.AIResponse{}).Where("id = ?", messageID).Updates(map[string]interface{}{"success": true, "used_at": readAt}).Error
}
