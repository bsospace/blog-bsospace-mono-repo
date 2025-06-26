package auth

import (
	handler "rag-searchbot-backend/api/v1/auth/me"
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/pkg/crypto"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RegisterRoutes กำหนดเส้นทาง API สำหรับ Authentication
func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, cache *cache.Service, logger *zap.Logger) {
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

	// สร้างกลุ่ม Route `/auth`
	authRoutes := router.Group("/auth")

	// auth middleware
	authMiddleware := middleware.AuthMiddleware(userService, cryptoService, cacheService, logger)

	// ใช้ Middleware ตรวจสอบ JWT
	authRoutes.Use(authMiddleware)
	{
		// Me Route
		authRoutes.GET("/me", handler.Me)
	}
}
