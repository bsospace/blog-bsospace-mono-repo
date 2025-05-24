package post

import (
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/post"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/pkg/crypto"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, cache *cache.Service) {
	// Inject dependencies
	repo := post.NewPostRepository(db)
	service := post.NewPostService(repo)
	handler := NewPostHandler(service)

	// สร้าง Repository ที่ใช้ GORM
	userRepository := user.NewRepository(db)

	// สร้าง Service ที่ใช้ Crypto
	crypto := crypto.NewCryptoService()

	// cryptoService
	cryptoService := crypto
	// สร้าง Service ที่ใช้ Repository
	userService := user.NewService(userRepository)

	// Cache Service
	cacheService := cache

	// auth middleware
	authMiddleware := middleware.AuthMiddleware(userService, cryptoService, cacheService)

	// Protected Category Routes
	postsRoutes := router.Group("/posts")
	{
		postsRoutes.GET("", handler.GetAll)
		// postsRoutes.GET("/:slug", handler.GetBySlug)
	}

	postsRoutes.Use(authMiddleware)
	{
		postsRoutes.POST("", handler.Create)
		postsRoutes.GET("/:short_slug", handler.GetByShortSlug)

	}
}
