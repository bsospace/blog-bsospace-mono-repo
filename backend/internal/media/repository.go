package media

import (
	"rag-searchbot-backend/internal/models"

	"github.com/google/uuid"
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

func (r *MediaRepository) DeleteByID(id string) error {
	return r.DB.Where("id = ?", id).Delete(&models.ImageUpload{}).Error
}

func (r *MediaRepository) MakeAsUsed(id uint, reason string) error {
	return r.DB.Model(&models.ImageUpload{}).Where("id = ?", id).Updates(map[string]interface{}{
		"IsUsed":     true,
		"UsedReason": reason,
	}).Error
}

// GetImagesByPostID ดึงรูปภาพทั้งหมดที่เชื่อมกับโพสต์นี้
func (m *MediaRepository) GetImagesByPostID(postID uuid.UUID) ([]models.ImageUpload, error) {
	var images []models.ImageUpload
	err := m.DB.Where("post_id = ?", postID).Find(&images).Error
	return images, err
}

// UpdateImageUsage บันทึกสถานะการใช้งานของรูปภาพ (is_used, used_at)
func (m *MediaRepository) UpdateImageUsage(image *models.ImageUpload) error {
	return m.DB.Save(image).Error
}

func (m *MediaRepository) DeleteImagesWhereUnused() error {
	return m.DB.
		Where("is_used = ?", false).
		Delete(&models.ImageUpload{}).Error
}
