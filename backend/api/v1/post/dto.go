package post

import (
	"rag-searchbot-backend/internal/models"
	"time"
)

type PostContentStructure struct {
	Type    string                 `json:"type"`
	Attrs   map[string]string      `json:"attrs,omitempty"`
	Marks   []map[string]string    `json:"marks,omitempty"`
	Text    string                 `json:"text,omitempty"`
	Content []PostContentStructure `json:"content,omitempty"`
}
type GetPublicPostBySlugAndUsernameResponse struct {
	ID          string     `json:"id"`
	Slug        string     `json:"slug"`
	ShortSlug   string     `json:"short_slug"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Thumbnail   string     `json:"thumbnail"`
	Content     string     `json:"content"`
	Published   bool       `json:"published"`
	PublishedAt time.Time  `json:"published_at"`
	Likes       int        `json:"likes"`
	Views       int        `json:"views"`
	ReadTime    int        `json:"read_time"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
	AuthorID    string     `json:"author_id"`
	AIChatOpen  bool       `json:"ai_chat_open"`
	AIReady     bool       `json:"ai_ready"`
	Author      struct {
		Avatar    string `json:"avatar"`
		Username  string `json:"username"`
		Bio       string `json:"bio"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"author"`
}

func MapGetPublicPostBySlugAndUsernameResponse(post *models.Post) *GetPublicPostBySlugAndUsernameResponse {
	if post == nil {
		return nil
	}

	dto := &GetPublicPostBySlugAndUsernameResponse{
		ID:          post.ID.String(),
		Slug:        post.Slug,
		ShortSlug:   post.ShortSlug,
		Title:       post.Title,
		Description: post.Description,
		Thumbnail:   post.Thumbnail,
		Content:     post.Content,
		Published:   post.Published,
		PublishedAt: *post.PublishedAt,
		Likes:       post.Likes,
		Views:       post.Views,
		ReadTime:    int(post.ReadTime),
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		DeletedAt:   &post.DeletedAt.Time,
		AIChatOpen:  post.AIChatOpen,
		AIReady:     post.AIReady,
	}

	dto.Author = struct {
		Avatar    string `json:"avatar"`
		Username  string `json:"username"`
		Bio       string `json:"bio"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}{
		Avatar:    post.Author.Avatar,
		Username:  post.Author.UserName,
		Bio:       post.Author.Bio,
		FirstName: post.Author.FirstName,
		LastName:  post.Author.LastName,
	}

	return dto
}
