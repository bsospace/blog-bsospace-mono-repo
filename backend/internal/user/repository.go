package user

import (
	"rag-searchbot-backend/internal/models"

	"gorm.io/gorm"
)

type RepositoryInterface interface {
	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)
	GetUsers() ([]models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByUsername(username string) (bool, error)
	GetUserProfileByUsername(username string) (*models.User, error)
	UpdateUser(user *models.User) error
}

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) RepositoryInterface {
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

// func (r *Repository) preloadUserRelations(db *gorm.DB) *gorm.DB {
// 	return db.
// 		Preload("Posts").
// 		Preload("Comments").
// 		Preload("AIUsageLogs").
// 		Preload("Notifications")
// }

func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.DB.
		Where("email = ?", email).
		Where("deleted_at IS NULL").
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByUsername(username string) (bool, error) {
	var user models.User
	err := r.DB.
		Where("username = ?", username).
		Where("deleted_at IS NULL").
		First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil // Username not found
	} else if err != nil {
		return false, err // Other error occurred
	}

	return true, nil
}

// GetUserProfileByUsername ค้นหา User Profile โดย Username
func (r *Repository) GetUserProfileByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.DB.
		Where("username = ?", username).
		Where("deleted_at IS NULL").
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update
func (r *Repository) UpdateUser(user *models.User) error {
	return r.DB.Model(&models.User{}).
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"username":   user.UserName,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"bio":        user.Bio,
			"avatar":     user.Avatar,
			"location":   user.Location,
			"website":    user.Website,
			"github":     user.GitHub,
			"twitter":    user.Twitter,
			"linkedin":   user.LinkedIn,
			"instagram":  user.Instagram,
			"facebook":   user.Facebook,
			"youtube":    user.YouTube,
			"discord":    user.Discord,
			"telegram":   user.Telegram,
			"new_user":   false,
		}).Error
}
