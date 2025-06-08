package post

import (
	"encoding/json"
	"errors"
	"math"
	"rag-searchbot-backend/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PostService struct {
	Repo *PostRepository
}

func NewPostService(repo *PostRepository) *PostService {
	return &PostService{Repo: repo}
}

type TagDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type CategoryDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (s *PostService) CreatePost(post CreatePostRequest, user *models.User) error {
	slug := post.ShortSlug + "-" + user.ID.String()

	// 1. Marshal content
	contentJSON, err := json.Marshal(post.Content)
	if err != nil {
		return err
	}

	// 2. check if a post with the same short slug already exists
	existingPost, err := s.Repo.GetByShortSlug(slug)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 3. if a post with the same short slug exists, update it
	if existingPost != nil {
		existingPost.Content = string(contentJSON)
		existingPost.Title = post.Title
		return s.Repo.Update(existingPost)
	}

	// 4. if no post with the same short slug exists, create a new one
	newPost := &models.Post{
		Slug:        slug,
		ShortSlug:   slug,
		Content:     string(contentJSON),
		Title:       post.Title,
		AuthorID:    user.ID,
		Published:   false,
		PublishedAt: nil,
	}

	return s.Repo.Create(newPost)
}

func (s *PostService) GetByShortSlug(shortSlug string) (*models.Post, error) {
	return s.Repo.GetByShortSlug(shortSlug)
}

/*
*
  - GetPosts retrieves a paginated list of posts.
  - @param c *gin.Context - The Gin context
  - @return *PostListResponse - The response containing the list of posts
  - @return error - An error if occurred
*/

func (s *PostService) GetPosts(c *gin.Context) (*PostListResponse, error) {
	limitStr := c.DefaultQuery("limit", "10")
	pageStr := c.DefaultQuery("page", "1")
	search := c.Query("search")

	if search == "" {
		search = " "
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	result, err := s.Repo.GetAll(limit, offset, search)
	if err != nil {
		return nil, err
	}

	var meta Meta

	meta.Total = result.Total
	meta.HasNextPage = result.HasNext
	meta.Page = result.Page
	meta.Limit = result.Limit
	meta.TotalPage = int(math.Ceil(float64(result.Total) / float64(limit)))

	var postDTOs []PostSummaryDTO
	for _, post := range result.Posts {
		postDTOs = append(postDTOs, MapPostToSummaryDTO(post))
	}

	return &PostListResponse{
		Posts: postDTOs,
		Meta:  meta,
	}, nil
}

func (s *PostService) GetPostByID(id string) (*models.Post, error) {
	post, err := s.Repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, nil
	}

	if post.Key != "" {
		return nil, nil
	}
	return post, nil
}

/*
* GetPostBySlug retrieves a post by its slug.
* @param slug string - The slug of the post
* @return *PostByIdResponse - The response containing the post
 */

func (s *PostService) GetPostBySlug(slug string) (*PostByIdResponse, error) {
	post, err := s.Repo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, nil
	}

	if post.Key != "" {
		return nil, nil
	}
	postDTO := MapPostToSummaryDTOWithContent(*post)

	return &PostByIdResponse{
		Post: postDTO,
	}, nil
}

func (s *PostService) UpdatePost(post *models.Post) error {
	return s.Repo.Update(post)
}

func (s *PostService) DeletePost(id string) error {
	return s.Repo.Delete(id)
}

type MyPostsResponseDTO struct {
	Posts []MyPostsDTO `json:"posts"`
}

func (s *PostService) MyPosts(user *models.User) (*MyPostsResponseDTO, error) {
	rawPosts, err := s.Repo.getMyPosts(user)
	if err != nil {
		return nil, err
	}

	var postDTOs []MyPostsDTO
	for _, post := range rawPosts {
		postDTOs = append(postDTOs, MapMyPostToSummaryDTO(*post))
	}

	return &MyPostsResponseDTO{
		postDTOs,
	}, nil
}

func (s *PostService) PublishPost(post *PublishPostRequestDTO, user *models.User, shortSlug string) error {

	shortSlug = shortSlug + "-" + user.ID.String()

	existingPost, err := s.Repo.GetByShortSlug(shortSlug)

	if err != nil {
		return err
	}

	if existingPost == nil {
		return errors.New("post not found")
	}

	// if existingPost.Published {
	// 	return errors.New("post is already published")
	// }

	if existingPost.AuthorID != user.ID {
		return errors.New("you are not the author of this post")
	}

	existingPost.Published = true
	now := time.Now()
	existingPost.Title = post.Title
	existingPost.PublishedAt = &now
	existingPost.Slug = post.Slug
	existingPost.Description = post.Description
	existingPost.Thumbnail = post.Thumbnail

	return s.Repo.Update(existingPost)
}

func (s *PostService) UnpublishPost(user *models.User, shortSlug string) error {

	shortSlug = shortSlug + "-" + user.ID.String()

	existingPost, err := s.Repo.GetByShortSlug(shortSlug)

	if err != nil {
		return err
	}

	if existingPost == nil {
		return errors.New("post not found")
	}

	if existingPost.AuthorID != user.ID {
		return errors.New("you are not the author of this post")
	}

	return s.Repo.UnpublishPost(existingPost)
}

func (s *PostService) GetPublicPostBySlugAndUsername(slug string, username string) (*models.Post, error) {
	post, err := s.Repo.GetPublicPostBySlugAndUsername(slug, username)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, nil
	}

	if post.Key != "" {
		return nil, nil
	}
	return post, nil
}
