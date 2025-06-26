package notification

import (
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/notification"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/internal/ws"
	"rag-searchbot-backend/pkg/crypto"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	db *gorm.DB, cache *cache.Service,
	logger *zap.Logger, asynqClient *asynq.Client,
	mux *asynq.ServeMux,
	socketManager *ws.Manager,
) {
	// Create Notification Repository
	repo := notification.NewRepository(db)
	// Create Notification Service
	notiService := notification.NewService(repo, socketManager)

	// Create Notification Handler
	handler := NewNotificationHandler(notiService)

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
	authMiddleware := middleware.AuthMiddleware(userService, cryptoService, cacheService, logger)

	// Protected Notification Routes
	notificationRoutes := router.Group("/notifications")
	notificationRoutes.Use(authMiddleware)
	{
		notificationRoutes.GET("", handler.GetNotificationsHandler)                          // Get all notifications for the user
		notificationRoutes.POST("/:id/mark-read", handler.MarkNotificationAsReadHandler)     // Mark all notifications as seen
		notificationRoutes.POST("/mark-all-read", handler.MarkAllNotificationsAsReadHandler) // Mark all notifications as seen
		notificationRoutes.DELETE("/:id/delete", handler.DeleteNotificationHandler)          // Delete

	}

}
