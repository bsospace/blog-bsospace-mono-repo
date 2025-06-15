package notification

import (
	"rag-searchbot-backend/internal/models"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CRUD: Create
func (r *Repository) Create(noti *models.Notification) (*models.Notification, error) {
	if err := r.db.Create(noti).Error; err != nil {
		return nil, err
	}
	return noti, nil
}
