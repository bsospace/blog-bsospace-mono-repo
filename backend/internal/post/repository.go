package post

import (
	"rag-searchbot-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Constants for popular posts configuration
const (
	MinPopularPostViews = 0 // Minimum view count required for a post to be considered popular
)

type PostRepositoryInterface interface {
	Create(post *models.Post) (string, error)
	GetAll(limit, offset int, search string) (*PostRepositoryQuery, error)
	GetByID(id string) (*models.Post, error)
	GetBySlug(slug string) (*models.Post, error)
	Update(post *models.Post) error
	GetMyPosts(user *models.User) ([]*models.Post, error)
	GetByShortSlug(shortSlug string) (*models.Post, error)
	GetPublicPostBySlugAndUsername(slug string, username string) (*models.Post, error)
	GetPublishedPostsByAuthor(username string, page, limit int) ([]models.Post, int64, error)
	PublishPost(post *models.Post) error
	UnpublishPost(post *models.Post) error
	DeletePost(post *models.Post) error
	GetEmbeddingByPostID(postID string) ([]models.Embedding, error)
	InsertEmbedding(post *models.Post, embedding models.Embedding) error
	UpdateEmbedding(post *models.Post, embedding models.Embedding) error
	DeleteEmbeddingsByPostID(postID string) error
	BulkInsertEmbeddings(post *models.Post, embeddings []models.Embedding) error
	RecordPostView(postID string, userID *string, fingerprint string, ipAddress, userAgent string) error
	GetPostViews(postID string) (int, error)
	GetPopularPosts(limit int) ([]models.Post, error)
}

type PostRepository struct {
	DB *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepositoryInterface {
	return &PostRepository{DB: db}
}

func (r *PostRepository) Create(post *models.Post) (string, error) {
	err := r.DB.Create(post).Error
	if err != nil {
		return "", err
	}
	return post.ID.String(), nil
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

	query := r.DB.
		Select("id", "slug", "title", "description", "thumbnail",
			"published", "published_at", "author_id", "likes",
			"views", "read_time", "ai_chat_open", "ai_ready").
		Where("published = ?", true).
		Where("deleted_at IS NULL").
		Where("published_at IS NOT NULL").
		Where("status = ?", models.PostPublished).
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "avatar")
		}).
		Preload("Tags").
		Preload("Categories").
		Order("published_at DESC").
		Limit(limit).
		Offset(offset)

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ?", searchTerm, searchTerm)
	}

	err := query.Find(&posts).Error
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
	query := r.DB.Model(&models.Post{}).
		Where("published = ?", true).
		Where("deleted_at IS NULL").
		Where("published_at IS NOT NULL")

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ?", searchTerm, searchTerm)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *PostRepository) GetByID(id string) (*models.Post, error) {
	var post models.Post

	err := r.DB.
		Select("id", "slug", "title", "content", "description",
			"thumbnail", "published", "published_at", "author_id",
			"likes", "views", "read_time", "ai_chat_open", "ai_ready").
		Where("deleted_at IS NULL").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "avatar")
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
			return db.Select("id", "username", "avatar")
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
	existing := &models.Post{}
	r.DB = r.DB.Debug()
	if err := r.DB.First(existing, "id = ?", post.ID).Error; err != nil {
		return err
	}

	updates := map[string]interface{}{}

	// เช็คค่าที่ส่งมาก่อนเพิ่มเข้า updates
	if post.Published != existing.Published {
		updates["published"] = post.Published
	}
	if post.AIChatOpen != existing.AIChatOpen {
		updates["ai_chat_open"] = post.AIChatOpen
	}
	if post.AIReady != existing.AIReady {
		updates["ai_ready"] = post.AIReady
	}

	// เปรียบเทียบทีละฟิลด์ ถ้า post ส่งค่ามาให้ (ไม่เป็น default) ก็ใช้ค่านั้น
	if post.PublishedAt != nil {
		updates["published_at"] = post.PublishedAt
	}

	if post.Title != "" {
		updates["title"] = post.Title
	}
	if post.Description != "" {
		updates["description"] = post.Description
	}
	if post.Thumbnail != "" {
		updates["thumbnail"] = post.Thumbnail
	}
	if post.Content != "" {
		updates["content"] = post.Content
	}
	if post.HTMLContent != nil && *post.HTMLContent != "" {
		updates["html_content"] = *post.HTMLContent
	}
	if post.Slug != "" {
		updates["slug"] = post.Slug
	}
	if post.ShortSlug != "" {
		updates["short_slug"] = post.ShortSlug
	}
	if len(post.Keywords) > 0 {
		updates["keywords"] = post.Keywords
	}
	if post.Example != "" {
		updates["example"] = post.Example
	}
	if post.ReadTime != 0 {
		updates["read_time"] = post.ReadTime
	}
	if post.Status != "" {
		updates["status"] = post.Status
	}
	if post.Key != "" {
		updates["key"] = post.Key
	}

	if len(updates) == 0 {
		return nil // ไม่มีฟิลด์ไหนเปลี่ยน
	}

	return r.DB.Model(post).Updates(updates).Error
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
		Select("id", "slug", "short_slug", "title", "content", "description", "thumbnail", "published", "status", "published_at", "author_id", "likes", "views", "read_time").
		Where("deleted_at IS NULL").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "avatar")
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
		Select("posts.id", "posts.slug", "posts.title", "posts.content",
			"posts.description", "posts.thumbnail", "posts.published", "posts.published_at",
			"posts.author_id", "posts.likes", "posts.views", "posts.read_time",
			"posts.created_at", "posts.updated_at", "posts.ai_chat_open", "posts.ai_ready").
		Where("posts.deleted_at IS NULL").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "avatar", "bio")
		}).
		Preload("Tags").
		Preload("Categories").
		Joins("JOIN users ON users.id = posts.author_id").
		Where("posts.slug = ? AND users.username = ? AND posts.published = ?", slug, username, true).
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
		"published":    post.Published,
		"published_at": post.PublishedAt,
		"status":       post.Status,
	}).Error

}

func (r *PostRepository) DeletePost(post *models.Post) error {
	return r.DB.Delete(&models.Post{}, "id = ?", post.ID).Error
}

func (r *PostRepository) GetEmbeddingByPostID(postID string) ([]models.Embedding, error) {
	var post models.Post
	err := r.DB.Preload("Embeddings").Where("id = ?", postID).First(&post).Error
	if err != nil {
		return nil, err
	}
	return post.Embeddings, nil
}

// insert embedding into the database

func (r *PostRepository) InsertEmbedding(post *models.Post, embedding models.Embedding) error {
	embedding.PostID = post.ID // ต้อง set foreign key ด้วย
	err := r.DB.Create(&embedding).Error
	if err != nil {
		return err
	}
	return nil
}

// update embedding in the database
func (r *PostRepository) UpdateEmbedding(post *models.Post, embedding models.Embedding) error {
	var existingEmbedding models.Embedding
	err := r.DB.Where("post_id = ? AND id = ?", post.ID, embedding.ID).First(&existingEmbedding).Error
	if err != nil {
		return err
	}

	existingEmbedding.Content = embedding.Content
	existingEmbedding.Vector = embedding.Vector

	return r.DB.Save(&existingEmbedding).Error
}

// DeleteEmbeddingsByPostID

func (r *PostRepository) DeleteEmbeddingsByPostID(postID string) error {
	return r.DB.
		Unscoped().
		Where("post_id = ?", postID).
		Delete(&models.Embedding{}).Error
}

func (r *PostRepository) BulkInsertEmbeddings(post *models.Post, embeddings []models.Embedding) error {
	if len(embeddings) == 0 {
		return nil
	}
	return r.DB.Create(&embeddings).Error
}

// RecordPostView บันทึก view ของ post
func (r *PostRepository) RecordPostView(postID string, userID *string, fingerprint string, ipAddress, userAgent string) error {
	// แปลง string เป็น uuid.UUID
	postUUID, err := uuid.Parse(postID)
	if err != nil {
		return err
	}

	var userUUID *uuid.UUID
	if userID != nil {
		parsedUserID, err := uuid.Parse(*userID)
		if err != nil {
			return err
		}
		userUUID = &parsedUserID
	}

	// ตรวจสอบว่า fingerprint นี้เคย view post นี้แล้วหรือไม่
	var existingView models.PostView
	err = r.DB.Where("post_id = ? AND fingerprint = ?", postUUID, fingerprint).First(&existingView).Error

	// ถ้าไม่เคย view ให้สร้างใหม่
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			postView := models.PostView{
				PostID:      postUUID,
				UserID:      userUUID,
				Fingerprint: fingerprint,
				IPAddress:   ipAddress,
				UserAgent:   userAgent,
			}

			// สร้าง PostView ใหม่
			if err := r.DB.Create(&postView).Error; err != nil {
				return err
			}

			// อัปเดต view count ในตาราง posts
			if err := r.DB.Model(&models.Post{}).Where("id = ?", postUUID).UpdateColumn("views", gorm.Expr("views + ?", 1)).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// GetPostViews ดึงจำนวน view ของ post
func (r *PostRepository) GetPostViews(postID string) (int, error) {
	var count int64
	err := r.DB.Model(&models.PostView{}).Where("post_id = ?", postID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetPublishedPostsByAuthor ดึงบทความที่ published โดย author คนหนึ่ง
func (r *PostRepository) GetPublishedPostsByAuthor(username string, page, limit int) ([]models.Post, int64, error) {
	var posts []models.Post
	offset := (page - 1) * limit

	// ดึง user ID จาก username
	var user models.User
	err := r.DB.Where("username = ? AND deleted_at IS NULL", username).First(&user).Error
	if err != nil {
		return nil, 0, err
	}

	// ดึงบทความที่ published โดย author คนนี้
	query := r.DB.
		Select("id", "slug", "title", "description", "thumbnail",
			"published", "published_at", "author_id", "likes",
			"views", "read_time", "ai_chat_open", "ai_ready").
		Where("author_id = ?", user.ID).
		Where("published = ?", true).
		Where("deleted_at IS NULL").
		Where("published_at IS NOT NULL").
		Where("status = ?", models.PostPublished).
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "avatar")
		}).
		Preload("Tags").
		Preload("Categories").
		Order("published_at DESC").
		Limit(limit).
		Offset(offset)

	err = query.Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	// นับจำนวนบทความทั้งหมด
	var total int64
	err = r.DB.Model(&models.Post{}).
		Where("author_id = ?", user.ID).
		Where("published = ?", true).
		Where("deleted_at IS NULL").
		Where("published_at IS NOT NULL").
		Where("status = ?", models.PostPublished).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPopularPosts ดึงบทความยอดนิยมตามจำนวน view
func (r *PostRepository) GetPopularPosts(limit int) ([]models.Post, error) {
	var posts []models.Post

	err := r.DB.
		Select("id", "slug", "title", "description", "thumbnail",
			"published", "published_at", "author_id", "likes",
			"views", "read_time", "ai_chat_open", "ai_ready").
		Where("published = ?", true).
		Where("deleted_at IS NULL").
		Where("published_at IS NOT NULL").
		Where("status = ?", models.PostPublished).
		Where("views > ?", MinPopularPostViews). // เฉพาะบทความที่มี view มากกว่า 0
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username", "avatar")
		}).
		Preload("Tags").
		Preload("Categories").
		Order("views DESC, published_at DESC"). // เรียงตาม view count จากมากไปน้อย, ถ้า view เท่ากันให้เรียงตามวันที่ publish
		Limit(limit).
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	return posts, nil
}
