package main

import (
	"log"
	"time"

	"rag-searchbot-backend/config"
	"rag-searchbot-backend/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

func SeedDatabase(db *gorm.DB) {
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. Create Users
		user := models.User{
			ID:       uuid.New(),
			Email:    "admin@blog.com",
			Password: "hashed_password",
			UserName: "admin",
			Avatar:   "/images/admin.png",
			Image:    "/images/admin.png",
			Bio:      "I write about tech and life.",
			Role:     models.AdminUser,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// 2. Tags
		tags := []models.Tag{
			{Name: "AI"},
			{Name: "Tech"},
			{Name: "Life"},
		}
		if err := tx.Create(&tags).Error; err != nil {
			return err
		}

		// 3. Categories
		categories := []models.Category{
			{Name: "Development"},
			{Name: "Lifestyle"},
		}
		if err := tx.Create(&categories).Error; err != nil {
			return err
		}

		// 4. Post
		post := models.Post{
			ID:          uuid.New(),
			Slug:        "first-blog-post",
			Title:       "Welcome to My Blog",
			Description: "This is the first blog post about starting a blog.",
			Content:     "Lorem ipsum dolor sit amet...",
			Published:   true,
			PublishedAt: ptrTime(time.Now()),
			Keywords:    []string{"blog", "start", "intro"},
			Likes:       10,
			Views:       200,
			ReadTime:    3.5,
			AuthorID:    user.ID,
		}
		if err := tx.Create(&post).Error; err != nil {
			return err
		}
		if err := tx.Model(&post).Association("Tags").Append(tags); err != nil {
			return err
		}
		if err := tx.Model(&post).Association("Categories").Append(categories); err != nil {
			return err
		}

		// 5. Comment
		comment := models.Comment{
			Content:  "Great post!",
			PostID:   post.ID,
			AuthorID: user.ID,
		}
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}

		// 6. Notification
		noti := models.Notification{
			Title:   "Welcome!",
			Content: "Thanks for joining the blog!",
			Link:    "/dashboard",
			UserID:  user.ID,
			Seen:    false,
		}
		if err := tx.Create(&noti).Error; err != nil {
			return err
		}

		// 7. Embedding
		embedding := models.Embedding{
			ID:      uuid.New(),
			PostID:  post.ID,
			Content: "Intro to blog post...",
			Vector:  pq.Float64Array{0.1, 0.2, 0.3, 0.4},
		}
		if err := tx.Create(&embedding).Error; err != nil {
			return err
		}

		// 8. AIUsageLog
		aiLog := models.AIUsageLog{
			UserID:    user.ID,
			Action:    "summarize",
			TokenUsed: 32,
			Success:   true,
		}
		if err := tx.Create(&aiLog).Error; err != nil {
			return err
		}

		// 9. AIResponse
		aiRes := models.AIResponse{
			UserID:      user.ID,
			PostID:      post.ID,
			EmbeddingID: embedding.ID,
			Prompt:      "What is this post about?",
			Response:    "It's an introductory blog post about blogging.",
			TokenUsed:   28,
			Success:     true,
		}
		if err := tx.Create(&aiRes).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatalf("[SEED] ❌ Seed failed: %v", err)
	}
	log.Println("[SEED] ✅ Seed data inserted successfully!")
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func main() {
	// Example usage
	db := config.ConnectDatabase()
	SeedDatabase(db)
}
