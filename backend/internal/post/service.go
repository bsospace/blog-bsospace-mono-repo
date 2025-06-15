package post

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/pkg/errs"
	"rag-searchbot-backend/pkg/logger"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PostService struct {
	Repo         PostRepositoryInterface
	MediaService *media.MediaService
	TaskEnqueuer *TaskEnqueuer
}

func NewPostService(repo PostRepositoryInterface, mediaRepo *media.MediaService, enqueuer *TaskEnqueuer) *PostService {
	return &PostService{Repo: repo, MediaService: mediaRepo, TaskEnqueuer: enqueuer}
}

type TagDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type CategoryDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

/**
* CreatePost creates a new post or updates an existing one if a post with the same short slug already exists.
* @param post CreatePostRequest - The request containing post details
* @param user *models.User - The user creating the post
* @return string - The ID of the created or updated post
* @return error - An error if occurred
* This function marshals the content of the post, checks if a post with the same short slug exists,
* updates the existing post if it does, or creates a new post if it doesn't.
* It also updates the image usage status based on the content of the post.
* If an error occurs during any of these operations, it returns the error.
* If the post is created successfully, it returns the ID of the post.
* If an existing post is updated, it returns the ID of the updated post.
* The short slug is combined with the user ID to ensure uniqueness.
**/

func (s *PostService) CreatePost(post CreatePostRequest, user *models.User) (string, error) {
	slug := post.ShortSlug + "-" + user.ID.String()

	// 1. Marshal content
	contentJSON, err := json.Marshal(post.Content)
	if err != nil {
		return "", err
	}

	// 2. check if a post with the same short slug already exists
	existingPost, err := s.Repo.GetByShortSlug(slug)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	// 3. if a post with the same short slug exists, update it
	if existingPost != nil {
		existingPost.Content = string(contentJSON)
		existingPost.Title = post.Title

		err = s.UpdateImageUsageStatus(existingPost, post.Content, "")
		if err != nil {
			return "", err
		}

		err = s.Repo.Update(existingPost)
		if err != nil {
			return "", err
		}

		return existingPost.ID.String(), nil
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

	postID, err := s.Repo.Create(newPost)
	if err != nil {
		return "", err
	}

	return postID, nil
}

/**
* GetByShortSlug retrieves a post by its short slug.
* @param shortSlug string - The short slug of the post
* @return *models.Post - The post if found, nil otherwise
* @return error - An error if occurred
* This function queries the repository for a post with the given short slug.
* If the post is found, it returns the post and nil error.
* If the post is not found, it returns nil and an error indicating that the post was not found.
* If an error occurs during the query, it returns nil and the error.
* The short slug is expected to be unique for each post, so this function should return at most one post.
* If the post is found, it returns the post with its ID, slug, title, content, description, thumbnail, published status,
* published date, author ID, likes, views, and read time.
* If the post is not found, it returns nil.
* If an error occurs during the query, it returns nil and the error.
**/

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

/**
* GetPostByID retrieves a post by its ID.
* @param id string - The ID of the post
* @return *models.Post - The post if found, nil otherwise
* @return error - An error if occurred
* This function queries the repository for a post with the given ID.
* If the post is found, it returns the post and nil error.
* If the post is not found, it returns nil and an error indicating that the post was not found.
* If an error occurs during the query, it returns nil and the error.
**/

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

/**
* UpdatePost updates an existing post.
* @param post *models.Post - The post to update
* @return error - An error if occurred
* This function updates the post in the repository.
* It expects the post to have a valid ID and all necessary fields populated.
* If the post is successfully updated, it returns nil.
* If an error occurs during the update, it returns the error.
* The post should have been created previously and should exist in the repository.
* The function does not check if the post exists before updating, so it is assumed that the caller has already verified this.
* The post can be updated with new content, title, description, thumbnail, and other fields as needed.
* The post's ID must be set to the ID of the post to be updated.
* If the post is not found in the repository, the update will fail and return an error.
**/

func (s *PostService) UpdatePost(post *models.Post) error {
	return s.Repo.Update(post)
}

type MyPostsResponseDTO struct {
	Posts []MyPostsDTO `json:"posts"`
}

/**
* MyPosts retrieves the posts created by the specified user.
* @param user *models.User - The user whose posts to retrieve
* @return *MyPostsResponseDTO - The response containing the user's posts
* @return error - An error if occurred
**/

func (s *PostService) MyPosts(user *models.User) (*MyPostsResponseDTO, error) {
	rawPosts, err := s.Repo.GetMyPosts(user)
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

/**
* PublishPost publishes a post if it exists and is not already published.
* @param post *PublishPostRequestDTO - The request containing post details
* @param user *models.User - The user publishing the post
* @param shortSlug string - The short slug of the post
* @return error - An error if occurred
* This function checks if a post with the given short slug exists and is not already published.
**/

func (s *PostService) PublishPost(post *PublishPostRequestDTO, user *models.User, shortSlug string) error {

	shortSlug = shortSlug + "-" + user.ID.String()

	existingPost, err := s.Repo.GetByShortSlug(shortSlug)

	if err != nil {
		return err
	}

	if existingPost == nil {
		return errors.New("post not found")
	}

	if existingPost.Published {
		return errors.New("post is already published")
	}

	if existingPost.AuthorID != user.ID {
		return errors.New("you are not the author of this post")
	}

	// Validate Slug is not duplicate
	existingPostBySlug, err := s.Repo.GetBySlug(post.Slug)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// If a post with the same slug exists, append user ID and a random UUID to the slug
	if existingPostBySlug != nil {
		if existingPostBySlug.ID != existingPost.ID {
			existingPost.Slug = post.Slug + "-" + "-" + uuid.New().String()[:8]
		}
	} else {
		// If no post with the same slug exists, use the provided slug
		existingPost.Slug = post.Slug
	}

	now := time.Now()
	existingPost.Title = post.Title
	existingPost.Description = post.Description
	existingPost.Thumbnail = post.Thumbnail
	existingPost.HTMLContent = post.HTMLContent

	cfg := config.LoadConfig()

	// log mode
	logger.Log.Info("Publishing post",
		zap.String("AppEnv", cfg.AppEnv),
	)

	if cfg.AppEnv == "debug" {

		existingPost.Published = false
		existingPost.Status = models.PostProcessing
		existingPost.PublishedAt = nil

		logger.Log.Info("Enqueuing post content for AI filtering ",
			zap.String("post_id", existingPost.ID.String()),
			zap.String("post_title", existingPost.Title),
			zap.String("author_id", existingPost.AuthorID.String()),
			zap.String("author_email", user.Email))
		_, err = s.TaskEnqueuer.EnqueueFilterPostContentByAI(existingPost, user)
		if err != nil {
			return err
		}
	} else {
		existingPost.Published = true
		existingPost.PublishedAt = &now
		existingPost.Status = models.PostPublished
	}

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

	existingPost.Published = false
	existingPost.PublishedAt = nil
	existingPost.Status = models.PostDraft

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

func (s *PostService) DeletePostByID(id string, user *models.User) error {
	existingPost, err := s.Repo.GetByID(id)
	if err != nil {
		return err
	}

	if existingPost == nil || existingPost.ID == uuid.Nil {
		return errs.ErrPostNotFound
	}
	if existingPost.AuthorID != user.ID {
		return errs.ErrUnauthorized
	}

	// Remove all images related to the post
	if err := s.MakeAllNotUsedImageStatus(existingPost); err != nil {
		return fmt.Errorf("failed to update image usage status: %w", err)
	}

	// Make thunbnail not used
	if err := s.UpdateThumbnailUsageStatus(existingPost, existingPost.Thumbnail); err != nil {
		return fmt.Errorf("failed to update thumbnail usage status: %w", err)
	}

	// Delete the post
	return s.Repo.DeletePost(existingPost)
}

// updateImageUsageStatus checks which images in content or thumbnail are used
// and updates the is_used flag accordingly for all images related to the post
func (s *PostService) UpdateImageUsageStatus(post *models.Post, content PostContentStructure, thumbnail string) error {
	existingPost, err := s.Repo.GetByID(post.ID.String())
	if err != nil {
		return err
	}

	//  ดึงภาพทั้งหมดในโพสต์
	images, err := s.MediaService.GetImagesByPostID(existingPost.ID)
	if err != nil {
		return err
	}

	// ดึง URL รูปจาก content tree
	contentImageURLs := ExtractImageURLsFromContent([]PostContentStructure{content})

	urlSet := make(map[string]bool)
	for _, url := range contentImageURLs {
		urlSet[url] = true
	}
	if thumbnail != "" {
		urlSet[thumbnail] = true
	}

	// ตรวจสอบการใช้งานรูป
	for _, img := range images {
		isUsed := urlSet[img.ImageURL]
		if img.IsUsed != isUsed {
			img.IsUsed = isUsed
			now := time.Now()
			img.UsedAt = &now

			if err := s.MediaService.UpdateImageUsage(&img); err != nil {
				return err
			}

			fmt.Printf("Updated image %s to used=%v\n", img.ImageURL, isUsed)
		}
	}

	return nil
}

// update thumbnail usage status
func (s *PostService) UpdateThumbnailUsageStatus(post *models.Post, thumbnail string) error {
	if thumbnail == "" {
		return nil // No thumbnail to update
	}

	// Get the existing post
	existingPost, err := s.Repo.GetByID(post.ID.String())
	if err != nil {
		return err
	}

	// Get all images related to the post
	images, err := s.MediaService.GetImagesByPostID(existingPost.ID)
	if err != nil {
		return err
	}

	// Check if the thumbnail is used in the images
	isUsed := false
	for _, img := range images {
		if img.ImageURL == thumbnail {
			isUsed = true
			break
		}
	}

	// get existing thumbnail image
	existingThumbnail, err := s.MediaService.GetImageByURL(thumbnail)

	if err != nil {
		return fmt.Errorf("failed to get existing thumbnail image: %w", err)
	}

	if existingThumbnail == nil {
		return fmt.Errorf("thumbnail image not found: %s", thumbnail)
	}

	// Update the thumbnail usage status

	existingThumbnail.IsUsed = isUsed
	existingThumbnail.UsedAt = &time.Time{}
	if err := s.MediaService.UpdateImageUsage(existingThumbnail); err != nil {
		return fmt.Errorf("failed to update thumbnail usage status: %w", err)
	}

	return nil
}

func ExtractImageURLsFromContent(content []PostContentStructure) []string {
	var urls []string

	var walk func(node PostContentStructure)
	walk = func(node PostContentStructure) {
		if node.Type == "image" && node.Attrs != nil {
			if srcRaw, ok := node.Attrs["src"]; ok {
				if src, ok := srcRaw.(string); ok {
					urls = append(urls, src)
				}
			}
		}
		for _, child := range node.Content {
			walk(child)
		}
	}

	for _, node := range content {
		walk(node)
	}

	return urls
}

func (s *PostService) MakeAllNotUsedImageStatus(post *models.Post) error {
	// Get all images related to the post
	images, err := s.MediaService.GetImagesByPostID(post.ID)
	if err != nil {
		return err
	}

	// Update all images to not used
	for _, img := range images {
		img.IsUsed = false
		img.UsedAt = nil

		if err := s.MediaService.UpdateImageUsage(&img); err != nil {
			return err
		}
	}

	return nil
}
