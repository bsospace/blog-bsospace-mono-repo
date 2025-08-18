package user

import (
	"rag-searchbot-backend/internal/container"
	"rag-searchbot-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, container *container.Container) {
	handler := NewUserHandler(container.UserService, container.PostService)
	authMiddleware := middleware.NewAuthMiddleware(
		container.UserService,
		container.CryptoService,
		container.CacheService,
		container.Log,
	)
	optionalAuthMiddleware := middleware.NewOptionalAuthMiddleware(
		container.UserService,
		container.CryptoService,
		container.CacheService,
		container.Log,
	)

	// Public routes with optional auth (can get user context if token provided)
	userRoutes := router.Group("/user")
	userRoutes.Use(optionalAuthMiddleware.Handler())
	{
		userRoutes.GET("/profile/:username", handler.GetUserProfile)
		userRoutes.GET("/profile/:username/posts", handler.GetUserProfileWithPosts)
		userRoutes.GET("/regions", handler.GetSupportedRegions)
	}

	// Protected routes (auth required)
	userRoutes.Use(authMiddleware.Handler())
	{
		userRoutes.GET("/existing-username", handler.GetExistingUsername)
		userRoutes.PUT("/update", handler.UpdateUser)
	}
}
