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

func (r *MediaRepository) GetByID(id uuid.UUID) (*models.ImageUpload, error) {
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
	subQuery := m.DB.
		Table("image_uploads").
		Select("file_id").
		Group("file_id").
		Having("COUNT(*) = 1")

	err := m.DB.
		Where("is_used = ?", false).
		Where("file_id IN (?)", subQuery).
		Delete(&models.ImageUpload{}).Error

	return err
}

func (m *MediaRepository) GetUnusedImages() ([]models.ImageUpload, error) {
	var images []models.ImageUpload
	err := m.DB.Where("is_used = ?", false).Find(&images).Error
	if err != nil {
		return nil, err
	}
	return images, nil
}

func (r *MediaRepository) GetByFileID(fileID string) (*models.ImageUpload, error) {
	var image models.ImageUpload
	if err := r.DB.Where("file_id = ?", fileID).First(&image).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *MediaRepository) FindUnusedWithUniqueFileID() ([]models.ImageUpload, error) {
	var results []models.ImageUpload

	subQuery := r.DB.
		Table("image_uploads").
		Select("file_id").
		Group("file_id").
		Having("COUNT(*) = 1")

	err := r.DB.
		Where("is_used = false").
		Where("file_id IN (?)", subQuery).
		Find(&results).Error

	return results, err
}
