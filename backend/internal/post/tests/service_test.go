package tests

import (
	"fmt"
	"testing"

	"mime/multipart"
	"rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/post"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Mock Repository

// Mock for PostRepositoryInterface

// ... existing code ...
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(post *models.Post) (string, error) {
	args := m.Called(post)
	return args.String(0), args.Error(1)
}
func (m *MockPostRepository) GetAll(limit, offset int, search string) (*post.PostRepositoryQuery, error) {
	args := m.Called(limit, offset, search)
	return args.Get(0).(*post.PostRepositoryQuery), args.Error(1)
}
func (m *MockPostRepository) GetByID(id string) (*models.Post, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}
func (m *MockPostRepository) GetBySlug(slug string) (*models.Post, error) {
	args := m.Called(slug)
	return args.Get(0).(*models.Post), args.Error(1)
}
func (m *MockPostRepository) Update(post *models.Post) error {
	args := m.Called(post)
	return args.Error(0)
}
func (m *MockPostRepository) GetMyPosts(user *models.User) ([]*models.Post, error) {
	args := m.Called(user)
	return args.Get(0).([]*models.Post), args.Error(1)
}
func (m *MockPostRepository) GetByShortSlug(shortSlug string) (*models.Post, error) {
	args := m.Called(shortSlug)
	return args.Get(0).(*models.Post), args.Error(1)
}
func (m *MockPostRepository) GetPublicPostBySlugAndUsername(slug string, username string) (*models.Post, error) {
	args := m.Called(slug, username)
	return args.Get(0).(*models.Post), args.Error(1)
}
func (m *MockPostRepository) PublishPost(post *models.Post) error {
	args := m.Called(post)
	return args.Error(0)
}
func (m *MockPostRepository) UnpublishPost(post *models.Post) error {
	args := m.Called(post)
	return args.Error(0)
}
func (m *MockPostRepository) DeletePost(post *models.Post) error {
	args := m.Called(post)
	return args.Error(0)
}
func (m *MockPostRepository) GetEmbeddingByPostID(postID string) ([]models.Embedding, error) {
	args := m.Called(postID)
	return args.Get(0).([]models.Embedding), args.Error(1)
}
func (m *MockPostRepository) InsertEmbedding(post *models.Post, embedding models.Embedding) error {
	args := m.Called(post, embedding)
	return args.Error(0)
}
func (m *MockPostRepository) UpdateEmbedding(post *models.Post, embedding models.Embedding) error {
	args := m.Called(post, embedding)
	return args.Error(0)
}
func (m *MockPostRepository) DeleteEmbeddingsByPostID(postID string) error {
	args := m.Called(postID)
	return args.Error(0)
}
func (m *MockPostRepository) BulkInsertEmbeddings(post *models.Post, embeddings []models.Embedding) error {
	args := m.Called(post, embeddings)
	return args.Error(0)
}

// เพิ่ม method ใหม่ที่ขาดหายไป
func (m *MockPostRepository) RecordPostView(postID string, userID *string, fingerprint string, ipAddress, userAgent string) error {
	args := m.Called(postID, userID, fingerprint, ipAddress, userAgent)
	return args.Error(0)
}

func (m *MockPostRepository) GetPostViews(postID string) (int, error) {
	args := m.Called(postID)
	return args.Int(0), args.Error(1)
}

// Add missing method GetPopularPosts
func (m *MockPostRepository) GetPopularPosts(limit int) ([]models.Post, error) {
	args := m.Called(limit)
	return args.Get(0).([]models.Post), args.Error(1)
}

// Add missing method GetPublishedPostsByAuthor
func (m *MockPostRepository) GetPublishedPostsByAuthor(username string, page, limit int) ([]models.Post, int64, error) {
	args := m.Called(username, page, limit)
	return args.Get(0).([]models.Post), args.Get(1).(int64), args.Error(2)
}

// Mock for MediaServiceInterface (minimal for this test)
type MockMediaService struct {
	mock.Mock
}

func (m *MockMediaService) CreateMedia(fileHeader *multipart.FileHeader, user *models.User, postID *uuid.UUID) (*models.ImageUpload, error) {
	return nil, nil
}
func (m *MockMediaService) DeleteFromChibisafe(image *models.ImageUpload) error {
	return nil
}
func (m *MockMediaService) GetImagesByPostID(postID uuid.UUID) ([]models.ImageUpload, error) {
	args := m.Called(postID)
	return args.Get(0).([]models.ImageUpload), args.Error(1)
}
func (m *MockMediaService) UpdateImageUsage(image *models.ImageUpload) error {
	args := m.Called(image)
	return args.Error(0)
}
func (m *MockMediaService) DeleteUnusedImages() error {
	return nil
}
func (m *MockMediaService) GetImageByURL(imageURL string) (*models.ImageUpload, error) {
	return nil, nil
}
func (m *MockMediaService) UploadToChibisafe(fileHeader *multipart.FileHeader) (media.ChibisafeResponse, error) {
	return media.ChibisafeResponse{}, nil
}

// Mock for TaskEnqueuer (minimal for this test)
type MockTaskEnqueuer struct{}

func TestCreatePost_UpdateExisting(t *testing.T) {
	repo := new(MockPostRepository)
	media := new(MockMediaService)
	enqueuer := &post.TaskEnqueuer{}

	service := post.NewPostService(repo, media, enqueuer)

	user := &models.User{ID: uuid.New()}
	content := post.PostContentStructure{Type: "paragraph", Text: "Hello"}
	postReq := post.CreatePostRequest{
		ShortSlug: "testslug",
		Title:     "Updated Title",
		Content:   content,
	}

	slug := postReq.ShortSlug + "-" + user.ID.String()
	existing := &models.Post{ID: uuid.New(), Slug: slug, ShortSlug: slug}

	repo.On("GetByShortSlug", slug).Return(existing, nil)
	repo.On("Update", mock.AnythingOfType("*models.Post")).Return(nil)
	// เพิ่ม expectation สำหรับ GetByID (service อาจเรียกใน UpdateImageUsageStatus)
	repo.On("GetByID", mock.Anything).Return(existing, nil)

	// Mock media service calls for UpdateImageUsageStatus
	media.On("GetImagesByPostID", existing.ID).Return([]models.ImageUpload{}, nil)

	id, err := service.CreatePost(postReq, user)
	assert.NoError(t, err)
	assert.Equal(t, existing.ID.String(), id)

	repo.AssertExpectations(t)
}

// Test case สำหรับ RecordPostView
func TestRecordPostView_Success(t *testing.T) {
	repo := new(MockPostRepository)
	media := new(MockMediaService)
	enqueuer := &post.TaskEnqueuer{}

	service := post.NewPostService(repo, media, enqueuer)

	user := &models.User{ID: uuid.New()}
	postID := uuid.New().String()
	fingerprint := "test-fingerprint-123"
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0 Test Browser"

	// Mock expectations
	repo.On("GetByID", postID).Return(&models.Post{ID: uuid.MustParse(postID)}, nil)
	repo.On("RecordPostView", postID, mock.AnythingOfType("*string"), fingerprint, ipAddress, userAgent).Return(nil)
	repo.On("GetPostViews", postID).Return(42, nil)

	// Test
	result, err := service.RecordPostView(postID, user, fingerprint, ipAddress, userAgent)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "View recorded successfully", result.Message)
	assert.Equal(t, 42, result.Views)

	repo.AssertExpectations(t)
}

// Test case สำหรับ RecordPostView โดยไม่มี user (anonymous)
func TestRecordPostView_AnonymousUser(t *testing.T) {
	repo := new(MockPostRepository)
	media := new(MockMediaService)
	enqueuer := &post.TaskEnqueuer{}

	service := post.NewPostService(repo, media, enqueuer)

	postID := uuid.New().String()
	fingerprint := "test-fingerprint-456"
	ipAddress := "192.168.1.2"
	userAgent := "Mozilla/5.0 Anonymous Browser"

	// Mock expectations
	repo.On("GetByID", postID).Return(&models.Post{ID: uuid.MustParse(postID)}, nil)
	repo.On("RecordPostView", postID, (*string)(nil), fingerprint, ipAddress, userAgent).Return(nil)
	repo.On("GetPostViews", postID).Return(15, nil)

	// Test
	result, err := service.RecordPostView(postID, nil, fingerprint, ipAddress, userAgent)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "View recorded successfully", result.Message)
	assert.Equal(t, 15, result.Views)

	repo.AssertExpectations(t)
}

// Test case สำหรับ RecordPostView เมื่อ post ไม่มีอยู่
func TestRecordPostView_PostNotFound(t *testing.T) {
	repo := new(MockPostRepository)
	media := new(MockMediaService)
	enqueuer := &post.TaskEnqueuer{}

	service := post.NewPostService(repo, media, enqueuer)

	postID := "non-existent-post-id"
	fingerprint := "test-fingerprint-789"
	ipAddress := "192.168.1.3"
	userAgent := "Mozilla/5.0 Test Browser"

	// Mock expectations - post not found
	repo.On("GetByID", postID).Return(nil, assert.AnError)

	// Test
	result, err := service.RecordPostView(postID, nil, fingerprint, ipAddress, userAgent)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)

	repo.AssertExpectations(t)
}

// Test case to verify that UpdateImageUsageStatus is called when creating a new post
func TestCreatePost_NewPost_UpdatesImageUsageStatus(t *testing.T) {
	repo := new(MockPostRepository)
	media := new(MockMediaService)
	enqueuer := &post.TaskEnqueuer{}

	service := post.NewPostService(repo, media, enqueuer)

	user := &models.User{ID: uuid.New()}
	imageURL := "https://example.com/image.png"

	// Create content with an image
	content := post.PostContentStructure{
		Type: "doc",
		Content: []post.PostContentStructure{
			{
				Type: "paragraph",
				Content: []post.PostContentStructure{
					{
						Type: "image",
						Attrs: map[string]interface{}{
							"src": imageURL,
						},
					},
				},
			},
		},
	}

	postReq := post.CreatePostRequest{
		ShortSlug: "newpost",
		Title:     "New Post with Image",
		Content:   content,
	}

	slug := postReq.ShortSlug + "-" + user.ID.String()
	postID := uuid.New().String()
	createdPost := &models.Post{
		ID:        uuid.MustParse(postID),
		Slug:      slug,
		ShortSlug: slug,
		Title:     postReq.Title,
		AuthorID:  user.ID,
	}

	// Mock image that should be marked as used
	imageUpload := models.ImageUpload{
		ID:       uuid.New(),
		ImageURL: imageURL,
		IsUsed:   false, // Initially not used
		PostID:   &createdPost.ID,
	}

	// Mock expectations
	repo.On("GetByShortSlug", slug).Return((*models.Post)(nil), gorm.ErrRecordNotFound) // No existing post
	repo.On("Create", mock.AnythingOfType("*models.Post")).Return(postID, nil)
	repo.On("GetByID", postID).Return(createdPost, nil)
	repo.On("DeleteEmbeddingsByPostID", postID).Return(nil)

	// Mock media service calls for UpdateImageUsageStatus
	media.On("GetImagesByPostID", createdPost.ID).Return([]models.ImageUpload{imageUpload}, nil)
	media.On("UpdateImageUsage", mock.MatchedBy(func(img *models.ImageUpload) bool {
		// Verify that the image is marked as used
		return img.ImageURL == imageURL && img.IsUsed == true
	})).Return(nil)

	// Test
	id, err := service.CreatePost(postReq, user)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, postID, id)

	repo.AssertExpectations(t)
	media.AssertExpectations(t)
}

// Test cases for GetPopularPosts functionality

// Test case สำหรับ GetPopularPosts สำเร็จ
func TestGetPopularPosts_Success(t *testing.T) {
	repo := new(MockPostRepository)
	media := new(MockMediaService)
	enqueuer := &post.TaskEnqueuer{}

	service := post.NewPostService(repo, media, enqueuer)

	limit := 4
	expectedPosts := []models.Post{
		{
			ID:          uuid.New(),
			Title:       "Most Popular Post",
			Slug:        "most-popular-post",
			Description: "This is the most popular post",
			Views:       1000,
			Likes:       150,
			Author: models.User{
				ID:       uuid.New(),
				UserName: "author1",
			},
		},
		{
			ID:          uuid.New(),
			Title:       "Second Popular Post",
			Slug:        "second-popular-post",
			Description: "This is the second most popular post",
			Views:       800,
			Likes:       120,
			Author: models.User{
				ID:       uuid.New(),
				UserName: "author2",
			},
		},
		{
			ID:          uuid.New(),
			Title:       "Third Popular Post",
			Slug:        "third-popular-post",
			Description: "This is the third most popular post",
			Views:       600,
			Likes:       90,
			Author: models.User{
				ID:       uuid.New(),
				UserName: "author3",
			},
		},
		{
			ID:          uuid.New(),
			Title:       "Fourth Popular Post",
			Slug:        "fourth-popular-post",
			Description: "This is the fourth most popular post",
			Views:       400,
			Likes:       60,
			Author: models.User{
				ID:       uuid.New(),
				UserName: "author4",
			},
		},
	}

	// Mock expectations
	repo.On("GetPopularPosts", limit).Return(expectedPosts, nil)

	// Test
	result, err := service.GetPopularPosts(limit)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 4, len(result.Posts))
	assert.Equal(t, int64(4), result.Meta.Total)
	assert.Equal(t, 1, result.Meta.Page)
	assert.Equal(t, limit, result.Meta.Limit)

	// Verify first post (most popular)
	firstPost := result.Posts[0]
	assert.Equal(t, "Most Popular Post", firstPost.Title)
	assert.Equal(t, 1000, firstPost.Views)
	assert.Equal(t, 150, firstPost.Likes)
	assert.Equal(t, "author1", firstPost.Author.UserName)

	// Verify last post (least popular)
	lastPost := result.Posts[3]
	assert.Equal(t, "Fourth Popular Post", lastPost.Title)
	assert.Equal(t, 400, lastPost.Views)
	assert.Equal(t, 60, lastPost.Likes)
	assert.Equal(t, "author4", lastPost.Author.UserName)

	repo.AssertExpectations(t)
}

// Test case สำหรับ GetPopularPosts เมื่อไม่มีโพสต์
func TestGetPopularPosts_EmptyResult(t *testing.T) {
	repo := new(MockPostRepository)
	media := new(MockMediaService)
	enqueuer := &post.TaskEnqueuer{}

	service := post.NewPostService(repo, media, enqueuer)

	limit := 4
	emptyPosts := []models.Post{}

	// Mock expectations
	repo.On("GetPopularPosts", limit).Return(emptyPosts, nil)

	// Test
	result, err := service.GetPopularPosts(limit)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result.Posts))
	assert.Equal(t, int64(0), result.Meta.Total)
	assert.Equal(t, 1, result.Meta.Page)
	assert.Equal(t, limit, result.Meta.Limit)

	repo.AssertExpectations(t)
}

// Test case สำหรับ GetPopularPosts เมื่อ repository error
func TestGetPopularPosts_RepositoryError(t *testing.T) {
	repo := new(MockPostRepository)
	media := new(MockMediaService)
	enqueuer := &post.TaskEnqueuer{}

	service := post.NewPostService(repo, media, enqueuer)

	limit := 4

	// Mock expectations - repository error
	repo.On("GetPopularPosts", limit).Return(([]models.Post)(nil), assert.AnError)

	// Test
	result, err := service.GetPopularPosts(limit)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, assert.AnError, err)

	repo.AssertExpectations(t)
}

// Test case สำหรับ GetPopularPosts ด้วย limit ที่แตกต่างกัน
func TestGetPopularPosts_DifferentLimits(t *testing.T) {
	repo := new(MockPostRepository)
	media := new(MockMediaService)
	enqueuer := &post.TaskEnqueuer{}

	service := post.NewPostService(repo, media, enqueuer)

	testCases := []struct {
		name     string
		limit    int
		expected int
	}{
		{"Limit 1", 1, 1},
		{"Limit 2", 2, 2},
		{"Limit 3", 3, 3},
		{"Limit 4", 4, 4},
		{"Limit 10", 10, 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock posts for this limit
			mockPosts := make([]models.Post, tc.expected)
			for i := 0; i < tc.expected; i++ {
				mockPosts[i] = models.Post{
					ID:    uuid.New(),
					Title: fmt.Sprintf("Post %d", i+1),
					Slug:  fmt.Sprintf("post-%d", i+1),
					Views: 1000 - (i * 100),
					Likes: 100 - (i * 10),
					Author: models.User{
						ID:       uuid.New(),
						UserName: fmt.Sprintf("author%d", i+1),
					},
				}
			}

			// Mock expectations
			repo.On("GetPopularPosts", tc.limit).Return(mockPosts, nil)

			// Test
			result, err := service.GetPopularPosts(tc.limit)

			// Assertions
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc.expected, len(result.Posts))
			assert.Equal(t, int64(tc.expected), result.Meta.Total)
			assert.Equal(t, 1, result.Meta.Page)
			assert.Equal(t, tc.limit, result.Meta.Limit)

			// Verify posts are ordered by views (descending)
			for i := 0; i < len(result.Posts)-1; i++ {
				assert.GreaterOrEqual(t, result.Posts[i].Views, result.Posts[i+1].Views)
			}
		})
	}

	repo.AssertExpectations(t)
}
