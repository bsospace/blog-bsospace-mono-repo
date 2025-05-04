package post

import (
	"rag-searchbot-backend/internal/models"
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
