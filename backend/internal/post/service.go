package post

import (
	"rag-searchbot-backend/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PostService struct {
	Repo *PostRepository
}

func NewPostService(repo *PostRepository) *PostService {
	return &PostService{Repo: repo}
}

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

type TagDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type CategoryDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type PostListResponse struct {
	Posts       []PostSummaryDTO `json:"posts"`
	Total       int64            `json:"total"`
	HasNextPage bool             `json:"hasNextPage"`
	Page        int              `json:"page"`
	Limit       int              `json:"limit"`
}

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

func (s *PostService) CreatePost(post *models.Post) error {
	return s.Repo.Create(post)
}

func (s *PostService) GetPosts(c *gin.Context) (*PostListResponse, error) {
	limitStr := c.DefaultQuery("limit", "10")
	pageStr := c.DefaultQuery("page", "1")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	result, err := s.Repo.GetAll(limit, offset)
	if err != nil {
		return nil, err
	}

	var postDTOs []PostSummaryDTO
	for _, post := range result.Posts {
		postDTOs = append(postDTOs, MapPostToSummaryDTO(post))
	}

	return &PostListResponse{
		Posts:       postDTOs,
		Total:       result.Total,
		HasNextPage: result.HasNext,
		Page:        result.Page,
		Limit:       result.Limit,
	}, nil
}

func (s *PostService) GetPostByID(id string) (*models.Post, error) {
	return s.Repo.GetByID(id)
}

func (s *PostService) UpdatePost(post *models.Post) error {
	return s.Repo.Update(post)
}

func (s *PostService) DeletePost(id string) error {
	return s.Repo.Delete(id)
}
