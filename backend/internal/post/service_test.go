package post_test

import (
	"encoding/json"
	"errors"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/post"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// -------- MOCK --------
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(p *models.Post) (string, error) {
	args := m.Called(p)
	return args.String(0), args.Error(1)
}
func (m *MockPostRepository) Update(p *models.Post) error {
	return m.Called(p).Error(0)
}
func (m *MockPostRepository) Delete(id string) error {
	return m.Called(id).Error(0)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}
func (m *MockPostRepository) GetByShortSlug(slug string) (*models.Post, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}
func (m *MockPostRepository) GetAll(limit, offset int, search string) (*post.PostRepositoryQuery, error) {
	args := m.Called(limit, offset, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*post.PostRepositoryQuery), args.Error(1)
}
func (m *MockPostRepository) GetMyPosts(user *models.User) ([]*models.Post, error) {
	args := m.Called(user)
	return args.Get(0).([]*models.Post), args.Error(1)
}
func (m *MockPostRepository) UnpublishPost(p *models.Post) error {
	return m.Called(p).Error(0)
}
func (m *MockPostRepository) PublishPost(p *models.Post) error {
	return m.Called(p).Error(0)
}
func (m *MockPostRepository) GetPublicPostBySlugAndUsername(slug, username string) (*models.Post, error) {
	args := m.Called(slug, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) DeletePost(post *models.Post) error {
	return m.Called(post).Error(0)
}

// -------- TESTS --------

func TestCreateNewPost(t *testing.T) {
	repo := new(MockPostRepository)
	svc := post.NewPostService(repo, nil)
	user := &models.User{ID: uuid.New()}
	content := post.PostContentStructure{
		Text: "hello",
	}

	slug := "testslug-" + user.ID.String()

	repo.On("GetByShortSlug", slug).Return(nil, nil)

	contentJSON, _ := json.Marshal(content)
	repo.On("Create", mock.MatchedBy(func(p *models.Post) bool {
		return p.Slug == slug && p.Content == string(contentJSON)
	})).Return("some-id", nil)

	_, err := svc.CreatePost(post.CreatePostRequest{
		ShortSlug: "testslug",
		Title:     "Hello World",
		Content:   content,
	}, user)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

type MockMediaService struct {
	mock.Mock
}

func (m *MockMediaService) GetImagesByPostID(postID uuid.UUID) ([]models.ImageUpload, error) {
	args := m.Called(postID)
	return args.Get(0).([]models.ImageUpload), args.Error(1)
}

func (m *MockMediaService) UpdateImageUsage(image *models.ImageUpload) error {
	args := m.Called(image)
	return args.Error(0)
}

// func TestUpdateExistingPost(t *testing.T) {
// 	repo := new(MockPostRepository)
// 	mediaSvc := new(MockMediaService)
// 	svc := post.NewPostService(repo, mediaSvc)

// 	user := &models.User{ID: uuid.New()}
// 	slug := "exist-" + user.ID.String()
// 	content := post.PostContentStructure{
// 		Text: "updated content",
// 	}
// 	contentJSON, _ := json.Marshal(content)

// 	// mock post
// 	existing := &models.Post{ID: uuid.New(), Slug: slug, AuthorID: user.ID}
// 	repo.On("GetByShortSlug", slug).Return(existing, nil)

// 	// must mock GetByID inside UpdateImageUsageStatus
// 	repo.On("GetByID", existing.ID.String()).Return(existing, nil)

// 	// must mock MediaService.GetImagesByPostID
// 	mediaSvc.On("GetImagesByPostID", existing.ID).Return([]models.ImageUpload{}, nil)

// 	// mock Update
// 	repo.On("Update", mock.MatchedBy(func(p *models.Post) bool {
// 		return p.Content == string(contentJSON)
// 	})).Return(nil)

// 	_, err := svc.CreatePost(post.CreatePostRequest{
// 		ShortSlug: "exist",
// 		Title:     "Updated",
// 		Content:   content,
// 	}, user)

// 	assert.NoError(t, err)
// 	repo.AssertExpectations(t)
// 	mediaSvc.AssertExpectations(t)
// }

func TestPublishPostSuccess(t *testing.T) {
	repo := new(MockPostRepository)
	svc := post.NewPostService(repo, nil)
	user := &models.User{ID: uuid.New()}
	shortSlug := "publish"
	fullSlug := shortSlug + "-" + user.ID.String()

	existing := &models.Post{ID: uuid.New(), AuthorID: user.ID, Published: false}
	repo.On("GetByShortSlug", fullSlug).Return(existing, nil)
	repo.On("GetBySlug", "slug-new").Return(nil, nil)
	repo.On("Update", mock.AnythingOfType("*models.Post")).Return(nil)

	err := svc.PublishPost(&post.PublishPostRequestDTO{
		Title:       "Post",
		Slug:        "slug-new",
		Description: "desc",
		Thumbnail:   "thumb.png",
	}, user, shortSlug)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUnpublishPostSuccess(t *testing.T) {
	repo := new(MockPostRepository)
	svc := post.NewPostService(repo, nil)
	user := &models.User{ID: uuid.New()}
	shortSlug := "post"
	fullSlug := shortSlug + "-" + user.ID.String()

	existing := &models.Post{AuthorID: user.ID}
	repo.On("GetByShortSlug", fullSlug).Return(existing, nil)
	repo.On("UnpublishPost", existing).Return(nil)

	err := svc.UnpublishPost(user, shortSlug)
	assert.NoError(t, err)
}

func TestGetPostByID(t *testing.T) {
	repo := new(MockPostRepository)
	svc := post.NewPostService(repo, nil)
	expected := &models.Post{ID: uuid.New(), Key: ""}
	repo.On("GetByID", "abc").Return(expected, nil)

	post, err := svc.GetPostByID("abc")
	assert.NoError(t, err)
	assert.Equal(t, expected, post)
}

func TestDeletePost(t *testing.T) {
	repo := new(MockPostRepository)
	svc := post.NewPostService(repo, nil)
	user := &models.User{ID: uuid.New()}
	post := &models.Post{ID: uuid.New(), AuthorID: user.ID}

	repo.On("GetByID", "abc").Return(post, nil)
	repo.On("DeletePost", post).Return(nil)

	err := svc.DeletePostByID("abc", user)
	assert.NoError(t, err)
}
func TestGetPostBySlug(t *testing.T) {
	repo := new(MockPostRepository)
	svc := post.NewPostService(repo, nil)

	t.Run("success", func(t *testing.T) {
		expectedPost := &models.Post{
			ID:      uuid.New(),
			Slug:    "test-slug",
			Title:   "Test Post",
			Content: "test content",
			Key:     "",
		}
		repo.On("GetBySlug", "test-slug").Return(expectedPost, nil)

		resp, err := svc.GetPostBySlug("test-slug")

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedPost.Title, resp.Post.Title)
	})

	t.Run("post not found", func(t *testing.T) {
		repo.On("GetBySlug", "nonexistent").Return(nil, nil)

		resp, err := svc.GetPostBySlug("nonexistent")

		assert.NoError(t, err)
		assert.Nil(t, resp)
	})

	t.Run("post with key should return nil", func(t *testing.T) {
		postWithKey := &models.Post{
			ID:  uuid.New(),
			Key: "some-key",
		}
		repo.On("GetBySlug", "with-key").Return(postWithKey, nil)

		resp, err := svc.GetPostBySlug("with-key")

		assert.NoError(t, err)
		assert.Nil(t, resp)
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("db error")
		repo.On("GetBySlug", "error").Return(nil, expectedErr)

		resp, err := svc.GetPostBySlug("error")

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})
}
