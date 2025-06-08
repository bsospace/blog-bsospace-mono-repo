package post_test

import (
	"encoding/json"
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

func (m *MockPostRepository) Create(p *models.Post) error {
	return m.Called(p).Error(0)
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
	})).Return(nil)

	err := svc.CreatePost(post.CreatePostRequest{
		ShortSlug: "testslug",
		Title:     "Hello World",
		Content:   content,
	}, user)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdateExistingPost(t *testing.T) {
	repo := new(MockPostRepository)
	svc := post.NewPostService(repo, nil)
	user := &models.User{ID: uuid.New()}
	slug := "exist-" + user.ID.String()
	content := post.PostContentStructure{
		Text: "updated content",
	}
	contentJSON, _ := json.Marshal(content)

	existing := &models.Post{Slug: slug, AuthorID: user.ID}
	repo.On("GetByShortSlug", slug).Return(existing, nil)
	repo.On("Update", mock.MatchedBy(func(p *models.Post) bool {
		return p.Content == string(contentJSON)
	})).Return(nil)

	err := svc.CreatePost(post.CreatePostRequest{
		ShortSlug: "exist",
		Title:     "Updated",
		Content:   content,
	}, user)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

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

	repo.On("Delete", "abc").Return(nil)

	err := svc.DeletePost("abc")
	assert.NoError(t, err)
}
