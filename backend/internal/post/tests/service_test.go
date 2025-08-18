package tests

import (
	"testing"

	"mime/multipart"
	"rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/post"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	return nil, nil
}
func (m *MockMediaService) UpdateImageUsage(image *models.ImageUpload) error {
	return nil
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
