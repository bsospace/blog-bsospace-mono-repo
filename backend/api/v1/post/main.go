package post

import (
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/post"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/pkg/crypto"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, cache *cache.Service) {
	// Inject dependencies

	mediaRepo := media.NewMediaRepository(db)
	mediaService := media.NewMediaService(mediaRepo)

	var postRepo post.PostRepositoryInterface = post.NewPostRepository(db)
	service := post.NewPostService(postRepo, mediaService)

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
	// Public routes (no authentication required)
	{
		postsRoutes.GET("", handler.GetAll)
		postsRoutes.GET("/public/:username/:slug", handler.GetPublicPostBySlugAndUsername)
	}

	postsRoutes.Use(authMiddleware)
	{
		postsRoutes.POST("", handler.Create)
		postsRoutes.GET("/:short_slug", handler.GetByShortSlug)
		postsRoutes.GET("/my-posts", handler.MyPost)
		postsRoutes.PUT("/publish/:short_slug", handler.Publish)
		postsRoutes.PUT("/unpublish/:short_slug", handler.Unpublish)
	}
}
