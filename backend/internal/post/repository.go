package post

import (
	"rag-searchbot-backend/internal/models"

	"gorm.io/gorm"
)

type PostRepository struct {
	DB *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{DB: db}
}

func (r *PostRepository) Create(post *models.Post) error {
	return r.DB.Create(post).Error
}

type PostRepositoryQuery struct {
	Limit   int           `json:"limit"`
	Total   int64         `json:"total"`
	HasNext bool          `json:"has_next"`
	Page    int           `json:"page"`
	Offset  int           `json:"offset"`
	Search  string        `json:"search"`
	Posts   []models.Post `json:"posts"`
}

func (r *PostRepository) GetAll(limit, offset int) (*PostRepositoryQuery, error) {
	var posts []models.Post
	err := r.DB.
		Select("id", "slug", "title", "description", "thumbnail", "published", "published_at", "author_id", "likes", "views", "read_time").
		Where("deleted_at IS NULL").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "user_name", "avatar")
		}).
		Preload("Tags").
		Preload("Categories").
		Order("published_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	total, err := r.getCount()
	if err != nil {
		return nil, err
	}

	hasNext := total > int64(offset+limit)

	result := &PostRepositoryQuery{
		Limit:   limit,
		Total:   total,
		HasNext: hasNext,
		Page:    offset/limit + 1,
		Offset:  offset,
		Posts:   posts,
	}

	return result, nil
}

func (r *PostRepository) getCount() (int64, error) {
	var count int64
	err := r.DB.Model(&models.Post{}).Where("deleted_at IS NULL").Count(&count).Error
	return count, err
}

func (r *PostRepository) GetByID(id string) (*models.Post, error) {
	var post models.Post
	err := r.DB.Preload("Author").Preload("Tags").Preload("Categories").First(&post, "id = ?", id).Error
	return &post, err
}

func (r *PostRepository) Update(post *models.Post) error {
	return r.DB.Save(post).Error
}

func (r *PostRepository) Delete(id string) error {
	return r.DB.Delete(&models.Post{}, "id = ?", id).Error
}
