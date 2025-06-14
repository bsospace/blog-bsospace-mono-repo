package post

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/pkg/errs"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostService struct {
	Repo         PostRepositoryInterface
	MediaService *media.MediaService
}

func NewPostService(repo PostRepositoryInterface, mediaRepo *media.MediaService) *PostService {
	return &PostService{Repo: repo, MediaService: mediaRepo}
}

type TagDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type CategoryDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

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

type MyPostsResponseDTO struct {
	Posts []MyPostsDTO `json:"posts"`
}

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

	existingPost.Published = true
	now := time.Now()
	existingPost.Title = post.Title
	existingPost.PublishedAt = &now
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
