package ws

import (
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/container"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/user"

	"github.com/gin-gonic/gin"
)

func StartWebSocketServer(router *gin.RouterGroup, container *container.Container) {
	socketAuthMiddleware := middleware.SocketAuthMiddleware(
		container.UserService.(*user.Service),
		container.CryptoService,
		container.CacheService.(*cache.Service),
		container.Log,
	)
	wsHandler := NewWebSocketHandler(container.SocketManager)
	wsGroup := router.Group("/ws").Use(socketAuthMiddleware)
	{
		wsGroup.GET("", wsHandler.HandleConnection)
		wsGroup.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})
	}
}
