package media

import (
	"rag-searchbot-backend/internal/container"
	"rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, container *container.Container) {
	handler := NewMediaHandler(container.MediaService.(*media.MediaService))
	authMiddleware := middleware.NewAuthMiddleware(
		container.UserService,
		container.CryptoService,
		container.CacheService,
		container.Log,
	)
	mediaRoutes := router.Group("/media")
	mediaRoutes.Use(authMiddleware.Handler())
	{
		mediaRoutes.POST("/upload", handler.UploadImageHandler)
	}
}
