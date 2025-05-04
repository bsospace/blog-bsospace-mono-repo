package post

import (
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/post"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, cache *cache.Service) {
	// Inject dependencies
	repo := post.NewPostRepository(db)
	service := post.NewPostService(repo)
	handler := NewPostHandler(service)

	// Protected Category Routes
	postsRoutes := router.Group("/posts")
	{
		postsRoutes.GET("", handler.GetAll)
		postsRoutes.GET("/:slug", handler.GetBySlug)
	}
}
