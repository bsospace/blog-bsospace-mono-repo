package auth

import (
	handler "rag-searchbot-backend/api/v1/auth/me"
	"rag-searchbot-backend/internal/container"
	"rag-searchbot-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes กำหนดเส้นทาง API สำหรับ Authentication
func RegisterRoutes(router *gin.RouterGroup, container *container.Container) {
	authRoutes := router.Group("/auth")
	authMiddleware := middleware.NewAuthMiddleware(
		container.UserService,
		container.CryptoService,
		container.CacheService,
		container.Log,
	)
	authRoutes.Use(authMiddleware.Handler())
	authRoutes.GET("/me", handler.Me)
}
