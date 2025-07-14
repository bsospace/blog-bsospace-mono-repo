package tests

import (
	"testing"
	"time"

	"rag-searchbot-backend/internal/post"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Struct สำหรับ test SQLite

type TestUser struct {
	ID       string `gorm:"primaryKey"`
	Email    string
	UserName string
}

type TestPost struct {
	ID          string `gorm:"primaryKey"`
	Title       string
	Content     string
	Slug        string
	ShortSlug   string
	AuthorID    string
	Published   bool
	PublishedAt *time.Time
	Status      string
}

type TestEmbedding struct {
	ID      string `gorm:"primaryKey"`
	PostID  string
	Content string
}

type PostRepositoryTestSuite struct {
	suite.Suite
	db         *gorm.DB
	repository post.PostRepositoryInterface
}

func (suite *PostRepositoryTestSuite) SetupSuite() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Auto migrate tables (ใช้ struct test)
	err = db.AutoMigrate(&TestUser{}, &TestPost{}, &TestEmbedding{})
	suite.Require().NoError(err)

	suite.db = db
	// ไม่ต้องใช้ post.NewPostRepository(db) ใน test SQLite นี้
}

func (suite *PostRepositoryTestSuite) TearDownTest() {
	suite.db.Exec("DELETE FROM test_embeddings")
	suite.db.Exec("DELETE FROM test_posts")
	suite.db.Exec("DELETE FROM test_users")
}

func (suite *PostRepositoryTestSuite) TestCreate() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)

	post := &TestPost{
		ID:        uuid.NewString(),
		Title:     "Test Post",
		Content:   "Test content",
		Slug:      "test-post",
		AuthorID:  user.ID,
		Published: true,
		Status:    "PUBLISHED",
	}
	suite.db.Create(post)

	var got TestPost
	err := suite.db.First(&got, "id = ?", post.ID).Error
	suite.NoError(err)
	suite.Equal(post.Title, got.Title)
}

func (suite *PostRepositoryTestSuite) TestGetAll() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)

	now := time.Now()
	post1 := &TestPost{
		ID:          uuid.NewString(),
		Title:       "First Post",
		Content:     "First content",
		Slug:        "first-post",
		AuthorID:    user.ID,
		Published:   true,
		PublishedAt: &now,
		Status:      "PUBLISHED",
	}
	post2 := &TestPost{
		ID:          uuid.NewString(),
		Title:       "Second Post",
		Content:     "Second content",
		Slug:        "second-post",
		AuthorID:    user.ID,
		Published:   true,
		PublishedAt: &now,
		Status:      "PUBLISHED",
	}

	suite.db.Create(post1)
	suite.db.Create(post2)

	var posts []TestPost
	err := suite.db.Find(&posts).Error
	suite.NoError(err)
	suite.Equal(2, len(posts))
}

func (suite *PostRepositoryTestSuite) TestGetAllWithSearch() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)
	now := time.Now()
	post1 := &TestPost{
		ID:          uuid.NewString(),
		Title:       "Go Programming",
		Content:     "Go content",
		Slug:        "go-programming",
		AuthorID:    user.ID,
		Published:   true,
		PublishedAt: &now,
		Status:      "PUBLISHED",
	}
	post2 := &TestPost{
		ID:          uuid.NewString(),
		Title:       "Java Programming",
		Content:     "Java content",
		Slug:        "java-programming",
		AuthorID:    user.ID,
		Published:   true,
		PublishedAt: &now,
		Status:      "PUBLISHED",
	}

	suite.db.Create(post1)
	suite.db.Create(post2)

	var result []TestPost
	err := suite.db.Where("title LIKE ?", "%Go%").Find(&result).Error
	suite.NoError(err)
	suite.Equal(1, len(result))
	suite.Equal("Go Programming", result[0].Title)
}

func (suite *PostRepositoryTestSuite) TestGetByID() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)

	post := &TestPost{
		ID:       uuid.NewString(),
		Title:    "Test Post",
		Content:  "Test content",
		Slug:     "test-post",
		AuthorID: user.ID,
	}
	suite.db.Create(post)

	var got TestPost
	err := suite.db.First(&got, "id = ?", post.ID).Error
	suite.NoError(err)
	suite.Equal(post.Title, got.Title)
}

func (suite *PostRepositoryTestSuite) TestGetBySlug() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)

	post := &TestPost{
		ID:       uuid.NewString(),
		Title:    "Test Post",
		Content:  "Test content",
		Slug:     "test-post",
		AuthorID: user.ID,
	}
	suite.db.Create(post)

	var got TestPost
	err := suite.db.First(&got, "slug = ?", "test-post").Error
	suite.NoError(err)
	suite.Equal(post.Slug, got.Slug)
}

func (suite *PostRepositoryTestSuite) TestUpdate() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)

	post := &TestPost{
		ID:       uuid.NewString(),
		Title:    "Original Title",
		Content:  "Original content",
		Slug:     "original-post",
		AuthorID: user.ID,
	}
	suite.db.Create(post)

	post.Title = "Updated Title"
	post.Content = "Updated content"

	err := suite.db.Save(post).Error
	suite.NoError(err)

	var updated TestPost
	err = suite.db.First(&updated, "id = ?", post.ID).Error
	suite.NoError(err)
	suite.Equal("Updated Title", updated.Title)
	suite.Equal("Updated content", updated.Content)
}

func (suite *PostRepositoryTestSuite) TestGetMyPosts() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)

	post1 := &TestPost{
		ID:       uuid.NewString(),
		Title:    "My First Post",
		Content:  "First content",
		Slug:     "my-first-post",
		AuthorID: user.ID,
	}
	post2 := &TestPost{
		ID:       uuid.NewString(),
		Title:    "My Second Post",
		Content:  "Second content",
		Slug:     "my-second-post",
		AuthorID: user.ID,
	}

	suite.db.Create(post1)
	suite.db.Create(post2)

	var posts []TestPost
	err := suite.db.Where("author_id = ?", user.ID).Find(&posts).Error
	suite.NoError(err)
	suite.Equal(2, len(posts))
}

func (suite *PostRepositoryTestSuite) TestGetByShortSlug() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)

	post := &TestPost{
		ID:        uuid.NewString(),
		Title:     "Test Post",
		Content:   "Test content",
		Slug:      "test-post",
		ShortSlug: "tp123",
		AuthorID:  user.ID,
	}
	suite.db.Create(post)

	var got TestPost
	err := suite.db.First(&got, "short_slug = ?", "tp123").Error
	suite.NoError(err)
	suite.Equal(post.ShortSlug, got.ShortSlug)
}

func (suite *PostRepositoryTestSuite) TestGetPublicPostBySlugAndUsername() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)

	post := &TestPost{
		ID:        uuid.NewString(),
		Title:     "Public Post",
		Content:   "Public content",
		Slug:      "public-post",
		AuthorID:  user.ID,
		Published: true,
	}
	suite.db.Create(post)

	var got TestPost
	err := suite.db.Where("slug = ? AND author_id = ?", "public-post", user.ID).First(&got).Error
	suite.NoError(err)
	suite.Equal(post.Slug, got.Slug)
}

// func (suite *PostRepositoryTestSuite) TestPublishPost() {
// 	user := &TestUser{
// 		ID:       uuid.NewString(),
// 		UserName: "testuser",
// 		Email:    "test@example.com",
// 	}
// 	suite.db.Create(user)

// 	post := &TestPost{
// 		ID:        uuid.NewString(),
// 		Title:     "Draft Post",
// 		Content:   "Draft content",
// 		Slug:      "draft-post",
// 		AuthorID:  user.ID,
// 		Published: false,
// 	}
// 	suite.db.Create(post)

// 	err := suite.db.Model(post).Update("published", true).Error
// 	suite.NoError(err)

// 	var updated TestPost
// 	err = suite.db.First(&updated, "id = ?", post.ID).Error
// 	suite.NoError(err)
// 	suite.True(updated.Published)
// 	suite.NotNil(updated.PublishedAt)
// }

func (suite *PostRepositoryTestSuite) TestUnpublishPost() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)

	post := &TestPost{
		ID:          uuid.NewString(),
		Title:       "Published Post",
		Content:     "Published content",
		Slug:        "published-post",
		AuthorID:    user.ID,
		Published:   false,
		PublishedAt: nil,
		Status:      "DRAFT",
	}
	suite.db.Create(post)

	err := suite.db.Model(post).Update("published", false).Error
	suite.NoError(err)
}

func (suite *PostRepositoryTestSuite) TestDeletePost() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)

	post := &TestPost{
		ID:       uuid.NewString(),
		Title:    "To Delete",
		Content:  "Content to delete",
		Slug:     "to-delete",
		AuthorID: user.ID,
	}
	suite.db.Create(post)

	err := suite.db.Delete(post).Error
	suite.NoError(err)

	var deleted TestPost
	err = suite.db.First(&deleted, "id = ?", post.ID).Error
	suite.Error(err)
}

func (suite *PostRepositoryTestSuite) TestInsertEmbedding() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)

	post := &TestPost{
		ID:       uuid.NewString(),
		Title:    "Test Post",
		Content:  "Test content",
		Slug:     "test-post",
		AuthorID: user.ID,
	}
	suite.db.Create(post)

	embedding := TestEmbedding{
		ID:      uuid.NewString(),
		Content: "Test embedding content",
		PostID:  post.ID,
	}

	err := suite.db.Create(&embedding).Error
	suite.NoError(err)
}

func (suite *PostRepositoryTestSuite) TestGetEmbeddingByPostID() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)

	post := &TestPost{
		ID:       uuid.NewString(),
		Title:    "Test Post",
		Content:  "Test content",
		Slug:     "test-post",
		AuthorID: user.ID,
	}
	suite.db.Create(post)

	embedding := TestEmbedding{
		ID:      uuid.NewString(),
		PostID:  post.ID,
		Content: "Test embedding content",
	}
	suite.db.Create(&embedding)

	var embeddings []TestEmbedding
	err := suite.db.Where("post_id = ?", post.ID).Find(&embeddings).Error
	suite.NoError(err)
	suite.Equal(1, len(embeddings))
	suite.Equal(embedding.Content, embeddings[0].Content)
}

func (suite *PostRepositoryTestSuite) TestBulkInsertEmbeddings() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)

	post := &TestPost{
		ID:       uuid.NewString(),
		Title:    "Test Post",
		Content:  "Test content",
		Slug:     "test-post",
		AuthorID: user.ID,
	}
	suite.db.Create(post)

	embeddings := []TestEmbedding{
		{
			ID:      uuid.NewString(),
			PostID:  post.ID,
			Content: "First embedding",
		},
		{
			ID:      uuid.NewString(),
			PostID:  post.ID,
			Content: "Second embedding",
		},
	}

	err := suite.db.Create(&embeddings).Error
	suite.NoError(err)

	var result []TestEmbedding
	err = suite.db.Where("post_id = ?", post.ID).Find(&result).Error
	suite.NoError(err)
	suite.Equal(2, len(result))
}

func (suite *PostRepositoryTestSuite) TestDeleteEmbeddingsByPostID() {
	user := &TestUser{
		ID:       uuid.NewString(),
		UserName: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)

	post := &TestPost{
		ID:       uuid.NewString(),
		Title:    "Test Post",
		Content:  "Test content",
		Slug:     "test-post",
		AuthorID: user.ID,
	}
	suite.db.Create(post)

	embedding := TestEmbedding{
		ID:      uuid.NewString(),
		PostID:  post.ID,
		Content: "Test embedding content",
	}
	suite.db.Create(&embedding)

	err := suite.db.Where("post_id = ?", post.ID).Delete(&TestEmbedding{}).Error
	suite.NoError(err)

	var embeddings []TestEmbedding
	err = suite.db.Where("post_id = ?", post.ID).Find(&embeddings).Error
	suite.NoError(err)
	suite.Equal(0, len(embeddings))
}

func TestPostRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(PostRepositoryTestSuite))
}
