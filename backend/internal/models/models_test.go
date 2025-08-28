package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserRole(t *testing.T) {
	tests := []struct {
		name     string
		role     UserRole
		expected string
	}{
		{
			name:     "normal user role",
			role:     NormalUser,
			expected: "NORMAL_USER",
		},
		{
			name:     "writer user role",
			role:     WriterUser,
			expected: "WRITER_USER",
		},
		{
			name:     "admin user role",
			role:     AdminUser,
			expected: "ADMIN_USER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.role))
		})
	}
}

func TestPostStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   PostStatus
		expected string
	}{
		{
			name:     "draft status",
			status:   PostDraft,
			expected: "DRAFT",
		},
		{
			name:     "processing status",
			status:   PostProcessing,
			expected: "PROCESSING",
		},
		{
			name:     "published status",
			status:   PostPublished,
			expected: "PUBLISHED",
		},
		{
			name:     "rejected status",
			status:   PostRejected,
			expected: "REJECTED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestBaseModel(t *testing.T) {
	now := time.Now()
	baseModel := BaseModel{
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, now, baseModel.CreatedAt)
	assert.Equal(t, now, baseModel.UpdatedAt)
	// DeletedAt is a GORM field that has a zero value, not nil
	assert.False(t, baseModel.DeletedAt.Valid)
}

func TestUser(t *testing.T) {
	userID := uuid.New()
	user := &User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		UserName:  "johndoe",
		Role:      NormalUser,
		NewUser:   true,
	}

	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
	assert.Equal(t, "johndoe", user.UserName)
	assert.Equal(t, NormalUser, user.Role)
	assert.True(t, user.NewUser)
}

func TestPost(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	post := &Post{
		ID:          postID,
		Slug:        "test-post",
		ShortSlug:   "tp",
		Title:       "Test Post",
		Description: "A test post",
		Content:     "This is test content",
		Published:   false,
		Status:      PostDraft,
		Likes:       0,
		Views:       0,
		ReadTime:    2.5,
		AuthorID:    authorID,
	}

	assert.Equal(t, postID, post.ID)
	assert.Equal(t, "test-post", post.Slug)
	assert.Equal(t, "tp", post.ShortSlug)
	assert.Equal(t, "Test Post", post.Title)
	assert.Equal(t, "A test post", post.Description)
	assert.Equal(t, "This is test content", post.Content)
	assert.False(t, post.Published)
	assert.Equal(t, PostDraft, post.Status)
	assert.Equal(t, 0, post.Likes)
	assert.Equal(t, 0, post.Views)
	assert.Equal(t, 2.5, post.ReadTime)
	assert.Equal(t, authorID, post.AuthorID)
}

func TestComment(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	comment := &Comment{
		ID:      uint(1),
		Content: "This is a test comment",
		PostID:  postID,
		AuthorID: authorID,
	}

	assert.Equal(t, uint(1), comment.ID)
	assert.Equal(t, "This is a test comment", comment.Content)
	assert.Equal(t, postID, comment.PostID)
	assert.Equal(t, authorID, comment.AuthorID)
}

func TestTag(t *testing.T) {
	tag := &Tag{
		ID:   uint(1),
		Name: "golang",
	}

	assert.Equal(t, uint(1), tag.ID)
	assert.Equal(t, "golang", tag.Name)
}

func TestCategory(t *testing.T) {
	category := &Category{
		ID:   uint(1),
		Name: "Programming",
	}

	assert.Equal(t, uint(1), category.ID)
	assert.Equal(t, "Programming", category.Name)
}

func TestEmbedding(t *testing.T) {
	postID := uuid.New()
	embedding := &Embedding{
		ID:      uuid.New(),
		PostID:  postID,
		Content: "This is embedding content",
		// Vector field will be set by the database
	}

	assert.Equal(t, postID, embedding.PostID)
	assert.Equal(t, "This is embedding content", embedding.Content)
	// Note: Vector field testing would require pgvector integration
}

func TestNotification(t *testing.T) {
	userID := uuid.New()
	notification := &Notification{
		ID:      uint(1),
		Title:   "New Comment",
		Even:    "comment_reply",
		Content: "Someone replied to your comment",
		Link:    "/post/123",
		Seen:    false,
		UserID:  userID,
	}

	assert.Equal(t, uint(1), notification.ID)
	assert.Equal(t, "New Comment", notification.Title)
	assert.Equal(t, "comment_reply", notification.Even)
	assert.Equal(t, "Someone replied to your comment", notification.Content)
	assert.Equal(t, "/post/123", notification.Link)
	assert.False(t, notification.Seen)
	assert.Equal(t, userID, notification.UserID)
}

func TestAIUsageLog(t *testing.T) {
	userID := uuid.New()
	usageLog := &AIUsageLog{
		ID:        uint(1),
		UserID:    userID,
		Action:    "chat_completion",
		TokenUsed: 150,
		Success:   true,
		Message:   "Successfully completed chat",
	}

	assert.Equal(t, uint(1), usageLog.ID)
	assert.Equal(t, userID, usageLog.UserID)
	assert.Equal(t, "chat_completion", usageLog.Action)
	assert.Equal(t, 150, usageLog.TokenUsed)
	assert.True(t, usageLog.Success)
	assert.Equal(t, "Successfully completed chat", usageLog.Message)
}

func TestAIResponse(t *testing.T) {
	userID := uuid.New()
	postID := uuid.New()
	aiResponse := &AIResponse{
		ID:        uint(1),
		UserID:    userID,
		PostID:    postID,
		Prompt:    "What is Go programming?",
		Response:  "Go is a programming language...",
		TokenUsed: 200,
		Success:   true,
		Model:     "gpt-3.5-turbo",
	}

	assert.Equal(t, uint(1), aiResponse.ID)
	assert.Equal(t, userID, aiResponse.UserID)
	assert.Equal(t, postID, aiResponse.PostID)
	assert.Equal(t, "What is Go programming?", aiResponse.Prompt)
	assert.Equal(t, "Go is a programming language...", aiResponse.Response)
	assert.Equal(t, 200, aiResponse.TokenUsed)
	assert.True(t, aiResponse.Success)
	assert.Equal(t, "gpt-3.5-turbo", aiResponse.Model)
}

func TestImageUpload(t *testing.T) {
	userID := uuid.New()
	postID := uuid.New()
	imageUpload := &ImageUpload{
		ID:         uuid.New(),
		UserID:     userID,
		PostID:     &postID,
		ImageURL:   "https://example.com/image.jpg",
		FileName:   "test-image.jpg",
		FileID:     "file123",
		Identifier: "test-image-123",
		IsUsed:     false,
		UsedReason: "blog",
	}

	assert.Equal(t, userID, imageUpload.UserID)
	assert.Equal(t, &postID, imageUpload.PostID)
	assert.Equal(t, "https://example.com/image.jpg", imageUpload.ImageURL)
	assert.Equal(t, "test-image.jpg", imageUpload.FileName)
	assert.Equal(t, "file123", imageUpload.FileID)
	assert.Equal(t, "test-image-123", imageUpload.Identifier)
	assert.False(t, imageUpload.IsUsed)
	assert.Equal(t, "blog", imageUpload.UsedReason)
}

func TestQueueTaskLog(t *testing.T) {
	userID := uuid.New()
	taskLog := &QueueTaskLog{
		ID:         uint(1),
		TaskID:     "task123",
		TaskType:   "FILTER_AI",
		RefID:      "post456",
		RefType:    "POST",
		Status:     "SUCCESS",
		Message:    "Task completed successfully",
		Duration:   1500,
		Payload:    `{"key": "value"}`,
		UserID:     userID,
	}

	assert.Equal(t, uint(1), taskLog.ID)
	assert.Equal(t, "task123", taskLog.TaskID)
	assert.Equal(t, "FILTER_AI", taskLog.TaskType)
	assert.Equal(t, "post456", taskLog.RefID)
	assert.Equal(t, "POST", taskLog.RefType)
	assert.Equal(t, "SUCCESS", taskLog.Status)
	assert.Equal(t, "Task completed successfully", taskLog.Message)
	assert.Equal(t, int64(1500), taskLog.Duration)
	assert.Equal(t, `{"key": "value"}`, taskLog.Payload)
	assert.Equal(t, userID, taskLog.UserID)
}

func TestPostView(t *testing.T) {
	postID := uuid.New()
	userID := uuid.New()
	postView := &PostView{
		ID:          uint(1),
		PostID:      postID,
		UserID:      &userID,
		Fingerprint: "fp123",
		IPAddress:   "192.168.1.1",
		UserAgent:   "Mozilla/5.0...",
	}

	assert.Equal(t, uint(1), postView.ID)
	assert.Equal(t, postID, postView.PostID)
	assert.Equal(t, &userID, postView.UserID)
	assert.Equal(t, "fp123", postView.Fingerprint)
	assert.Equal(t, "192.168.1.1", postView.IPAddress)
	assert.Equal(t, "Mozilla/5.0...", postView.UserAgent)
}
