package user

import (
	"rag-searchbot-backend/internal/container"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/user"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, container *container.Container) {
	handler := NewUserHandler(container.UserService.(*user.Service))
	authMiddleware := middleware.NewAuthMiddleware(
		container.UserService,
		container.CryptoService,
		container.CacheService,
		container.Log,
	)
	userRoutes := router.Group("/user")
	userRoutes.Use(authMiddleware.Handler())
	{
		userRoutes.GET("/check-username", handler.GetExistingUsername)
		userRoutes.PUT("/update", handler.UpdateUser)
	}
}
