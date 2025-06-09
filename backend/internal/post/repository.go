package post

import (
	"rag-searchbot-backend/internal/models"

	"gorm.io/gorm"
)

type PostRepositoryInterface interface {
	Create(post *models.Post) error
	GetAll(limit, offset int, search string) (*PostRepositoryQuery, error)
	GetByID(id string) (*models.Post, error)
	GetBySlug(slug string) (*models.Post, error)
	Update(post *models.Post) error
	GetMyPosts(user *models.User) ([]*models.Post, error)
	GetByShortSlug(shortSlug string) (*models.Post, error)
	GetPublicPostBySlugAndUsername(slug string, username string) (*models.Post, error)
	PublishPost(post *models.Post) error
	UnpublishPost(post *models.Post) error
	DeletePost(post *models.Post) error
}

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

func (r *PostRepository) GetAll(limit, offset int, search string) (*PostRepositoryQuery, error) {
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
		Where("published = ?", true).
		Where("deleted_at IS NULL").
		Where("published_at IS NOT NULL").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	total, err := r.getCount(search)
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

func (r *PostRepository) getCount(search string) (int64, error) {
	var count int64
	err := r.DB.Model(&models.Post{}).
		Where("title LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%").
		Where("published = ?", true).
		Where("deleted_at IS NULL").
		Where("published_at IS NOT NULL").
		Count(&count).Error
	return count, err
}

func (r *PostRepository) GetByID(id string) (*models.Post, error) {
	var post models.Post

	err := r.DB.
		Select("id", "slug", "title", "content", "description", "thumbnail", "published", "published_at", "author_id", "likes", "views", "read_time").
		Where("deleted_at IS NULL").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "user_name", "avatar")
		}).
		Preload("Tags").
		Preload("Categories").
		Where("id = ?", id).
		First(&post).Error
	if err != nil {
		return nil, err
	}

	return &post, err
}

func (r *PostRepository) GetBySlug(slug string) (*models.Post, error) {
	var post models.Post

	err := r.DB.
		Select("id", "slug", "title", "content", "description", "thumbnail", "published", "published_at", "author_id", "likes", "views", "read_time").
		Where("deleted_at IS NULL").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "user_name", "avatar")
		}).
		Preload("Tags").
		Preload("Categories").
		Where("slug = ?", slug).
		First(&post).Error
	if err != nil {
		return nil, err
	}

	return &post, err
}

func (r *PostRepository) Update(post *models.Post) error {
	existinPost := &models.Post{}
	err := r.DB.First(existinPost, "id = ?", post.ID).Error
	if err != nil {
		return err
	}
	return r.DB.Model(existinPost).Updates(post).Error
}

func (r *PostRepository) GetMyPosts(user *models.User) ([]*models.Post, error) {
	var posts []*models.Post
	err := r.DB.
		Where("author_id = ? AND deleted_at IS NULL", user.ID).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepository) GetByShortSlug(shortSlug string) (*models.Post, error) {
	var post models.Post

	err := r.DB.
		Select("id", "slug", "title", "content", "description", "thumbnail", "published", "published_at", "author_id", "likes", "views", "read_time").
		Where("deleted_at IS NULL").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "user_name", "avatar")
		}).
		Preload("Tags").
		Preload("Categories").
		Where("short_slug = ?", shortSlug).
		First(&post).Error
	if err != nil {
		return nil, err
	}

	return &post, err
}

func (r *PostRepository) GetPublicPostBySlugAndUsername(slug string, username string) (*models.Post, error) {
	var post models.Post

	err := r.DB.
		Select("posts.id", "posts.slug", "posts.title", "posts.content", "posts.description", "posts.thumbnail", "posts.published", "posts.published_at", "posts.author_id", "posts.likes", "posts.views", "posts.read_time", "posts.created_at", "posts.updated_at").
		Where("posts.deleted_at IS NULL").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "user_name", "avatar", "bio")
		}).
		Preload("Tags").
		Preload("Categories").
		Joins("JOIN users ON users.id = posts.author_id").
		Where("posts.slug = ? AND users.user_name = ? AND posts.published = ?", slug, username, true).
		First(&post).Error
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (r *PostRepository) PublishPost(post *models.Post) error {
	// Ensure the post is not already published
	if post.Published {
		return nil // or return an error if you prefer
	}

	post.Published = true
	post.PublishedAt = &post.CreatedAt

	return r.DB.Save(post).Error
}

func (r *PostRepository) UnpublishPost(post *models.Post) error {
	// Ensure the post is published before unpublishing
	return r.DB.Model(post).Updates(map[string]interface{}{
		"published":    false,
		"published_at": nil,
	}).Error

}

func (r *PostRepository) DeletePost(post *models.Post) error {
	return r.DB.Unscoped().Delete(&models.Post{}, "id = ?", post.ID).Error
}
