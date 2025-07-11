package ai

import (
	"rag-searchbot-backend/internal/ai"
	"rag-searchbot-backend/internal/container"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/notification"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, container *container.Container) {
	// Repository
	postRepo := container.PostRepo
	aiTaskEnqueuer := ai.NewTaskEnqueuer(container.AsynqClient)
	aiRepo := ai.NewAIRepository(container.DB)
	aiService := ai.NewAIService(postRepo, aiTaskEnqueuer, aiRepo)
	handler := NewAIHandler(aiService, postRepo, container.Log)
	authMiddleware := middleware.NewAuthMiddleware(
		container.UserService,
		container.CryptoService, // now correct type from internal/crypto
		container.CacheService,
		container.Log,
	)
	notificationService := container.NotificationService
	container.AsynqMux.HandleFunc(ai.TaskTypeEmbedPost, ai.NewEmbedPostWorkerHandler(ai.EmbedPostWorker{
		Logger:      container.Log,
		PostRepo:    postRepo,
		NotiService: notificationService.(*notification.NotificationService),
	}))
	aiRoutes := router.Group("/ai")
	aiRoutes.Use(authMiddleware.Handler())
	{
		aiRoutes.POST("/:post_id/on", handler.OpenAIMode)
		aiRoutes.POST("/:post_id/off", handler.DisableOpenAIMode)
		aiRoutes.POST("/:post_id/chat", handler.Chat)
	}
}
