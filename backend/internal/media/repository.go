package media

import (
	"rag-searchbot-backend/internal/models"

	"gorm.io/gorm"
)

type MediaRepository struct {
	DB *gorm.DB
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{DB: db}
}

func (r *MediaRepository) Create(media *models.ImageUpload) error {
	if err := r.DB.Create(media).Error; err != nil {
		return err
	}
	return r.DB.Preload("User").First(media, "id = ?", media.ID).Error
}

func (r *MediaRepository) GetByID(id uint) (*models.ImageUpload, error) {
	var media models.ImageUpload
	err := r.DB.Where("id = ?", id).First(&media).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}
