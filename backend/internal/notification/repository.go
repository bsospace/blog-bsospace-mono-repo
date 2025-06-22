package notification

import (
	"fmt"
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

// Get by user
func (r *Repository) GetByUser(userID string, limit int, page int) ([]models.Notification, error) {
	var notifications []models.Notification
	r.db = r.db.Debug()
	query := r.db.Where("user_id = ?", userID)

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	fmt.Printf("Fetching notifications for user %s with limit %d and offset %d\n", userID, limit, offset)

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

// CountByUser
func (r *Repository) CountByUser(userID string) (int64, error) {
	var count int64
	if err := r.db.Model(&models.Notification{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// make as read
func (r *Repository) MarkAsRead(notiID uint) error {
	if err := r.db.Model(&models.Notification{}).Where("id = ?", notiID).Updates(map[string]interface{}{
		"seen":    true,
		"seen_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}).Error; err != nil {
		return err
	}
	return nil
}

// mark all as read
func (r *Repository) MarkAllAsRead(userID string) error {
	if err := r.db.Model(&models.Notification{}).Where("user_id = ? AND seen = ?", userID, false).Updates(map[string]interface{}{
		"seen":    true,
		"seen_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}).Error; err != nil {
		return err
	}
	return nil
}

// get by ID
func (r *Repository) GetByID(notiID uint, user *models.User) (*models.Notification, error) {
	var noti models.Notification
	if err := r.db.Where("id = ? AND user_id = ?", notiID, user.ID).First(&noti).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, err
	}
	return &noti, nil
}

// delete by ID
func (r *Repository) DeleteByID(notiID uint, user *models.User) error {
	if err := r.db.Where("id = ? AND user_id = ?", notiID, user.ID).Delete(&models.Notification{}).Error; err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}
	return nil
}
