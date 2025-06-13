package user

import (
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/user"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, cache *cache.Service) {
	// สร้าง Repository ที่ใช้ GORM
	userRepository := user.NewRepository(db)

	// สร้าง Service ที่ใช้ Repository
	userService := user.NewService(userRepository)

	// สร้าง UserHandler
	userHandler := NewUserHandler(userService)

	// กำหนดเส้นทางสำหรับ User
	userRoutes := router.Group("/user")
	{
		// ตรวจสอบว่า username มีอยู่หรือไม่
		userRoutes.GET("/check-username", userHandler.GetExistingUsername)
	}
}
