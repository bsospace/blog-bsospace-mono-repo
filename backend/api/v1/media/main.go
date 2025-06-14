package media

import (
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/pkg/crypto"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, cache *cache.Service) {
	// Inject dependencies
	repo := media.NewMediaRepository(db)
	service := media.NewMediaService(repo)
	handler := NewMediaHandler(service)
	// สร้าง Repository ที่ใช้ GORM
	userRepository := user.NewRepository(db)

	// สร้าง Service ที่ใช้ Crypto
	crypto := crypto.NewCryptoService()

	// cryptoService
	cryptoService := crypto
	// สร้าง Service ที่ใช้ Repository
	userService := user.NewService(userRepository, cache)

	// Cache Service
	cacheService := cache

	// auth middleware
	authMiddleware := middleware.AuthMiddleware(userService, cryptoService, cacheService)

	// Protected Media Routes
	mediaRoutes := router.Group("/media")
	mediaRoutes.Use(authMiddleware)
	{
		mediaRoutes.POST("/upload", handler.UploadImageHandler)
	}
}
