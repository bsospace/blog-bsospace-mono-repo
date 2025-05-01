package user

import (
	"rag-searchbot-backend/internal/models"

	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// CreateUser
func (r *Repository) CreateUser(user *models.User) error {
	return r.DB.Create(user).Error
}

// GetUserByID ค้นหา User โดย ID
func (r *Repository) GetUserByID(id string) (*models.User, error) {
	var user models.User
	err := r.DB.First(&user, "id = ?", id).Error
	return &user, err
}

// GetUsers ดึง Users ทั้งหมด
func (r *Repository) GetUsers() ([]models.User, error) {
	var users []models.User
	err := r.DB.Find(&users).Error
	return users, err
}

func (r *Repository) preloadUserRelations(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Posts").
		Preload("Comments").
		Preload("AIUsageLogs").
		Preload("Notifications")
}

func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.preloadUserRelations(r.DB).
		Where("email = ?", email).
		Where("deleted_at IS NULL").
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
