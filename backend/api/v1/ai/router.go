package ai

import (
	"rag-searchbot-backend/internal/ai"
	"rag-searchbot-backend/internal/awsbedrock"
	"rag-searchbot-backend/internal/container"
	"rag-searchbot-backend/internal/llm"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/notification"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

func RegisterRoutes(router *gin.RouterGroup, container *container.Container, mux *asynq.ServeMux) {
	// Repository
	postRepo := container.PostRepo
	aiTaskEnqueuer := ai.NewTaskEnqueuer(container.AsynqClient)
	aiRepo := ai.NewAIRepository(container.DB)

	bedrockClient, err := awsbedrock.NewBedrockClient(*container.Env) // Instantiate Bedrock client
	if err != nil {
		container.Log.Fatal("Failed to create Bedrock client", zap.Error(err))
	}
	llmClient := llm.NewBedrockLLM(bedrockClient) // Instantiate LLM client

	aiContentClassifier := ai.NewAgentIntentClassifier(container.Log, postRepo, llmClient)
	aiService := ai.NewAIService(postRepo, aiTaskEnqueuer, aiRepo, aiContentClassifier, llmClient)
	agentToolWebSearch := ai.NewAgentToolWebSearchService(container.Log, postRepo, container.Env)
	handler := NewAIHandler(aiService, aiContentClassifier, postRepo, container.Log, agentToolWebSearch, llmClient)

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
