package main

import (
	"fmt"
	"log"
	"time"

	"rag-searchbot-backend/config"
	"rag-searchbot-backend/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// SeedDatabase เป็นฟังก์ชันสำหรับการ seed ข้อมูลเริ่มต้นลงในฐานข้อมูล
func SeedDatabase(db *gorm.DB) {
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. Create User
		user := models.User{
			ID:       uuid.New(),
			Email:    "admin@blog.com",
			UserName: "admin",
			Avatar:   "https://png.pngtree.com/png-vector/20220817/ourmid/pngtree-women-cartoon-avatar-in-flat-style-png-image_6110776.png",
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

		// 4. First Blog Post
		firstPost := models.Post{
			ID:          uuid.New(),
			Slug:        "first-blog-post",
			Title:       "Welcome to My Blog",
			Description: "This is the first blog post about starting a blog.",
			Content:     "Lorem ipsum dolor sit amet...",
			Thumbnail:   "https://developers.elementor.com/docs/assets/img/elementor-placeholder-image.png",
			Published:   true,
			PublishedAt: ptrTime(time.Now()),
			Keywords:    []string{"blog", "start", "intro"},
			Likes:       10,
			Views:       200,
			ReadTime:    3.5,
			AuthorID:    user.ID,
		}
		if err := tx.Create(&firstPost).Error; err != nil {
			return err
		}
		if err := tx.Model(&firstPost).Association("Tags").Append(tags); err != nil {
			return err
		}
		if err := tx.Model(&firstPost).Association("Categories").Append(categories); err != nil {
			return err
		}

		// 5. First Comment
		comment := models.Comment{
			Content:  "Great post!",
			PostID:   firstPost.ID,
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
			PostID:  firstPost.ID,
			Content: "Intro to blog post...",
			Vector:  pq.Float64Array{0.1, 0.2, 0.3, 0.4},
		}
		if err := tx.Create(&embedding).Error; err != nil {
			return err
		}

		// 8. AI Usage Log
		aiLog := models.AIUsageLog{
			UserID:    user.ID,
			Action:    "summarize",
			TokenUsed: 32,
			Success:   true,
		}
		if err := tx.Create(&aiLog).Error; err != nil {
			return err
		}

		// 9. AI Response
		aiRes := models.AIResponse{
			UserID:      user.ID,
			PostID:      firstPost.ID,
			EmbeddingID: embedding.ID,
			Prompt:      "What is this post about?",
			Response:    "It's an introductory blog post about blogging.",
			TokenUsed:   28,
			Success:     true,
		}
		if err := tx.Create(&aiRes).Error; err != nil {
			return err
		}

		// 10. Generate 22 More Blog Posts
		sampleTitles := []string{
			"Understanding LLMs: The Future of Language Models",
			"Getting Started with Docker for Developers",
			"10 VSCode Extensions to Boost Productivity",
			"What is Vector Search and Why It Matters",
			"Deploying Your App with Kubernetes in 10 Minutes",
			"How I Learned Go in 30 Days",
			"The Rise of AI in Daily Life",
			"Database Design: Tips for Scalable Systems",
			"Monorepo vs Polyrepo: What You Need to Know",
			"Integrating Redis for Real-Time Features",
			"How Search Engines Use Embeddings",
			"Working with REST APIs in Go",
			"Best Practices for Writing Technical Blogs",
			"React vs Svelte: A 2025 Comparison",
			"Making Your Blog SEO Friendly",
			"PostgreSQL Tricks Every Dev Should Know",
			"Using Git Effectively in Team Projects",
			"Monitoring with Prometheus and Grafana",
			"Optimizing Web Performance in 2025",
			"Build a Medium Clone with React and Go",
			"Why TypeScript is a Game Changer",
			"Automating CI/CD with Jenkins",
		}

		for i, title := range sampleTitles {
			slug := fmt.Sprintf("blog-post-%d", i+1)
			desc := fmt.Sprintf("This article covers the topic: %s.", title)
			content := fmt.Sprintf("## %s\n\nHere we dive into details about '%s'.", title, title)

			post := models.Post{
				ID:          uuid.New(),
				Slug:        slug,
				Title:       title,
				Description: desc,
				Content:     content,
				Thumbnail:   fmt.Sprintf("https://picsum.photos/seed/%d/800/400", i+1),
				Published:   true,
				PublishedAt: ptrTime(time.Now().Add(time.Duration(i) * time.Hour)),
				Keywords:    []string{"blog", "tech", "post"},
				Likes:       10 + i*2,
				Views:       200 + i*40,
				ReadTime:    3.5 + float64(i%5),
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

			// Add a comment for each post
			comment := models.Comment{
				Content:  fmt.Sprintf("I enjoyed reading: %s!", title),
				PostID:   post.ID,
				AuthorID: user.ID,
			}
			if err := tx.Create(&comment).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Fatalf("[SEED] ❌ Seed failed: %v", err)
	}
	log.Println("[SEED] ✅ Seed data inserted successfully!")
}

// ptrTime เป็นฟังก์ชันช่วยในการสร้าง pointer ไปยังเวลา
func ptrTime(t time.Time) *time.Time {
	return &t
}

// main ฟังก์ชันหลัก
func main() {
	// เชื่อมต่อฐานข้อมูล
	db := config.ConnectDatabase()
	// เรียกใช้ SeedDatabase เพื่อลงข้อมูล
	SeedDatabase(db)
}
