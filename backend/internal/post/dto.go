package post

import (
	"rag-searchbot-backend/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

/**
 * TagDTO represents a tag data transfer object.
 * It contains the ID and name of the tag.
**/

type PostSummaryDTO struct {
	ID          uuid.UUID  `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Thumbnail   string     `json:"thumbnail,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	Views       int        `json:"views"`
	Likes       int        `json:"likes"`
	ReadTime    float64    `json:"read_time"`
	Status      string     `json:"status"`
	Author      struct {
		ID       uuid.UUID `json:"id"`
		UserName string    `json:"username"`
		Avatar   string    `json:"avatar"`
	} `json:"author"`
	Tags       []TagDTO      `json:"tags,omitempty"`
	Categories []CategoryDTO `json:"categories,omitempty"`
}

/**
 * PostListResponse represents the response structure for a list of posts.
 * It contains a slice of PostSummaryDTO, total number of posts, pagination info.
**/
type PostListResponse struct {
	Posts []PostSummaryDTO `json:"posts"`
	Meta  Meta             `json:"meta"`
}

type Meta struct {
	Total       int64 `json:"total"`
	HasNextPage bool  `json:"hasNextPage"`
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	TotalPage   int   `json:"totalPage"`
}

/**
 * TagDTO represents a tag data transfer object.
 * It contains the ID and name of the tag.
**/

func MapPostToSummaryDTO(post models.Post) PostSummaryDTO {
	dto := PostSummaryDTO{
		ID:          post.ID,
		Slug:        post.Slug,
		Title:       post.Title,
		Description: post.Description,
		Thumbnail:   post.Thumbnail,
		PublishedAt: post.PublishedAt,
		Views:       post.Views,
		Likes:       post.Likes,
		ReadTime:    post.ReadTime,
		Status:      string(post.Status),
		Author: struct {
			ID       uuid.UUID `json:"id"`
			UserName string    `json:"username"`
			Avatar   string    `json:"avatar"`
		}{
			ID:       post.Author.ID,
			UserName: post.Author.UserName,
			Avatar:   post.Author.Avatar,
		},
	}

	for _, t := range post.Tags {
		dto.Tags = append(dto.Tags, TagDTO{ID: t.ID, Name: t.Name})
	}
	for _, c := range post.Categories {
		dto.Categories = append(dto.Categories, CategoryDTO{ID: c.ID, Name: c.Name})
	}

	return dto
}

type PostByIdResponse struct {
	Post PostByIdDTO `json:"post"`
}

type PostByIdDTO struct {
	ID          uuid.UUID  `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Thumbnail   string     `json:"thumbnail,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	Content     string     `json:"content"`
	Views       int        `json:"views"`
	Likes       int        `json:"likes"`
	ReadTime    float64    `json:"read_time"`
	Author      struct {
		ID       uuid.UUID `json:"id"`
		UserName string    `json:"username"`
		Avatar   string    `json:"avatar"`
	} `json:"author"`
	Tags       []TagDTO      `json:"tags,omitempty"`
	Categories []CategoryDTO `json:"categories,omitempty"`
}

func MapPostToSummaryDTOWithContent(post models.Post) PostByIdDTO {
	dto := PostByIdDTO{
		ID:          post.ID,
		Slug:        post.Slug,
		Title:       post.Title,
		Description: post.Description,
		Thumbnail:   post.Thumbnail,
		PublishedAt: post.PublishedAt,
		Content:     post.Content,
		Views:       post.Views,
		Likes:       post.Likes,
		ReadTime:    post.ReadTime,
		Author: struct {
			ID       uuid.UUID `json:"id"`
			UserName string    `json:"username"`
			Avatar   string    `json:"avatar"`
		}{
			ID:       post.Author.ID,
			UserName: post.Author.UserName,
			Avatar:   post.Author.Avatar,
		},
	}

	for _, t := range post.Tags {
		dto.Tags = append(dto.Tags, TagDTO{ID: t.ID, Name: t.Name})
	}
	for _, c := range post.Categories {
		dto.Categories = append(dto.Categories, CategoryDTO{ID: c.ID, Name: c.Name})
	}

	return dto
}

type PostContentStructure struct {
	Type    string                 `json:"type"`
	Content []PostContentStructure `json:"content,omitempty"`
	Attrs   map[string]interface{} `json:"attrs,omitempty"`
	Text    string                 `json:"text,omitempty"`
	Marks   []Mark                 `json:"marks,omitempty"`
}

type Mark struct {
	Type string `json:"type"`
}

func GroupByType(data []PostContentStructure) map[string][]PostContentStructure {
	grouped := make(map[string][]PostContentStructure)
	for _, item := range data {
		grouped[item.Type] = append(grouped[item.Type], item)
	}
	return grouped
}

type CreatePostRequest struct {
	ShortSlug string               `json:"short_slug" binding:"required"`
	Content   PostContentStructure `json:"content" binding:"required"`
	Title     string               `json:"title" binding:"required"`
}

type MyPostsDTO struct {
	ID          uuid.UUID  `json:"id"`
	Slug        string     `json:"slug"`
	ShortSlug   string     `json:"short_slug"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Thumbnail   string     `json:"thumbnail,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	Published   bool       `json:"published"`
	Status      string     `json:"status"`
	Views       int        `json:"views"`
	Likes       int        `json:"likes"`
	ReadTime    float64    `json:"read_time"`
	CreatedAt   time.Time  `json:"created_at"`
	AIChatOpen  bool       `json:"ai_chat_open"`
	AIReady     bool       `json:"ai_ready"`
}

func MapMyPostToSummaryDTO(post models.Post) MyPostsDTO {

	// spilt short slug
	shortSlugParts := strings.Split(post.ShortSlug, "-")
	shortSlug := post.ShortSlug

	if len(shortSlugParts) > 0 {
		shortSlug = shortSlugParts[0]
	}

	dto := MyPostsDTO{
		ID:          post.ID,
		Slug:        post.Slug,
		ShortSlug:   shortSlug,
		Title:       post.Title,
		Description: post.Description,
		Thumbnail:   post.Thumbnail,
		PublishedAt: post.PublishedAt,
		Status:      string(post.Status),
		Views:       post.Views,
		Likes:       post.Likes,
		ReadTime:    post.ReadTime,
		CreatedAt:   post.CreatedAt,
		AIChatOpen:  post.AIChatOpen,
		AIReady:     post.AIReady,
	}

	return dto
}

type PublishPostRequestDTO struct {
	Slug        string   `json:"slug"`
	Keywords    []string `json:"keywords"`
	Categories  []string `json:"categories"`
	Tags        []string `json:"tags"`
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Thumbnail   string   `json:"thumbnail"`
	HTMLContent *string  `json:"html_content"`
}
