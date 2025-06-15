package ws

import (
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/internal/ws"
	"rag-searchbot-backend/pkg/crypto"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func StartWebSocketServer(router *gin.RouterGroup, db *gorm.DB, cache *cache.Service, logger *zap.Logger, asynqClient *asynq.Client, mux *asynq.ServeMux, socketManager *ws.Manager) {

	// สร้าง Repository ที่ใช้ GORM
	userRepository := user.NewRepository(db)

	// สร้าง Service ที่ใช้ Repository
	userService := user.NewService(userRepository, cache)

	// สร้าง Service ที่ใช้ Crypto
	crypto := crypto.NewCryptoService()

	// cryptoService
	cryptoService := crypto
	// Cache Service
	cacheService := cache

	// auth middleware
	socketAuthMiddleware := middleware.SocketAuthMiddleware(userService, cryptoService, cacheService, logger)

	// Create a new WebSocket handler
	wsHandler := NewWebSocketHandler(socketManager)

	// Register the WebSocket route
	wsGroup := router.Group("/ws").Use(socketAuthMiddleware)
	{
		wsGroup.GET("", wsHandler.HandleConnection)
		wsGroup.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})
	}

}
