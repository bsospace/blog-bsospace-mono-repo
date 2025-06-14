package user

import (
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/pkg/crypto"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, cache *cache.Service) {
	// สร้าง Repository ที่ใช้ GORM
	userRepository := user.NewRepository(db)

	// สร้าง Service ที่ใช้ Repository
	userService := user.NewService(userRepository, cache)

	// สร้าง UserHandler
	userHandler := NewUserHandler(userService)

	// สร้าง Service ที่ใช้ Crypto
	crypto := crypto.NewCryptoService()

	// cryptoService
	cryptoService := crypto
	// Cache Service
	cacheService := cache

	// auth middleware
	authMiddleware := middleware.AuthMiddleware(userService, cryptoService, cacheService)

	// กำหนดเส้นทางสำหรับ User
	userRoutes := router.Group("/user")

	userRoutes.Use(authMiddleware)
	{
		// ตรวจสอบว่า username มีอยู่หรือไม่
		userRoutes.GET("/check-username", userHandler.GetExistingUsername)
		// อัพเดตข้อมูลผู้ใช้
		userRoutes.PUT("/update", userHandler.UpdateUser)
	}
}
