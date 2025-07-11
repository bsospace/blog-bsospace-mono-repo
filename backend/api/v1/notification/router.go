package notification

import (
	"rag-searchbot-backend/internal/container"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/notification"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, container *container.Container) {
	notiService := container.NotificationService
	handler := NewNotificationHandler(notiService.(*notification.NotificationService))
	authMiddleware := middleware.NewAuthMiddleware(
		container.UserService,
		container.CryptoService,
		container.CacheService,
		container.Log,
	)
	notificationRoutes := router.Group("/notifications")
	notificationRoutes.Use(authMiddleware.Handler())
	{
		notificationRoutes.GET("", handler.GetNotificationsHandler)
		notificationRoutes.POST(":id/mark-read", handler.MarkNotificationAsReadHandler)
		notificationRoutes.POST("/mark-all-read", handler.MarkAllNotificationsAsReadHandler)
		notificationRoutes.DELETE(":id/delete", handler.DeleteNotificationHandler)
	}
}
