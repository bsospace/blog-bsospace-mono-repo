package ai

import (
	"rag-searchbot-backend/internal/ai"
	"rag-searchbot-backend/internal/container"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/notification"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

func RegisterRoutes(router *gin.RouterGroup, container *container.Container, mux *asynq.ServeMux) {
	// Repository
	postRepo := container.PostRepo
	aiTaskEnqueuer := ai.NewTaskEnqueuer(container.AsynqClient)
	aiRepo := ai.NewAIRepository(container.DB)
	aiService := ai.NewAIService(postRepo, aiTaskEnqueuer, aiRepo)
	aiContentClassifier := ai.NewAgentIntentClassifier(container.Log, postRepo)
	agentToolWebSearch := ai.NewAgentToolWebSearchService(container.Log, postRepo, container.Env)
	handler := NewAIHandler(aiService, aiContentClassifier, postRepo, container.Log, agentToolWebSearch)

	authMiddleware := middleware.NewAuthMiddleware(
		container.UserService,
		container.CryptoService, // now correct type from internal/crypto
		container.CacheService,
		container.Log,
	)

	// container.AsynqMux.HandleFunc(ai.TaskTypeEmbedPost, ai.NewEmbedPostWorkerHandler(ai.EmbedPostWorker{
	// 	Logger:      container.Log,
	// 	PostRepo:    postRepo,
	// 	NotiService: notificationService.(*notification.NotificationService),
	// }))

	// Register AI task handlers
	mux.HandleFunc(ai.TaskTypeEmbedPost, ai.NewEmbedPostWorkerHandler(ai.EmbedPostWorker{
		Logger:      container.Log,
		PostRepo:    postRepo,
		NotiService: container.NotificationService.(*notification.NotificationService),
	}))

	aiRoutes := router.Group("/ai")
	aiRoutes.Use(authMiddleware.Handler())
	{
		aiRoutes.POST("/:post_id/on", handler.OpenAIMode)
		aiRoutes.POST("/:post_id/off", handler.DisableOpenAIMode)
		aiRoutes.POST("/:post_id/chat", handler.Chat)
		aiRoutes.GET("/:post_id/chats", handler.GetChatsByPost)
	}
}
