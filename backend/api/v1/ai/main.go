package ai

import (
	"rag-searchbot-backend/internal/ai"
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/notification"
	"rag-searchbot-backend/internal/post"
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
	// Repository
	postRepo := post.NewPostRepository(db)

	taskEnqueuer := ai.NewTaskEnqueuer(asynqClient)

	// AI Service
	aiService := ai.NewAIService(postRepo, taskEnqueuer)

	// Handler
	handler := NewAIHandler(aiService, postRepo, logger)

	// Middleware
	cryptoService := crypto.NewCryptoService()
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo, cache)
	authMiddleware := middleware.AuthMiddleware(userService, cryptoService, cache, logger)

	// notification service
	notificationRepo := notification.NewRepository(db)                              // Assuming you have a WebSocket manager
	notificationService := notification.NewService(notificationRepo, socketManager) // Assuming you have a WebSocket manager
	// Register AI worker
	// Register AI worker with full dependency injection
	mux.HandleFunc(ai.TaskTypeEmbedPost, ai.NewEmbedPostWorkerHandler(ai.EmbedPostWorker{
		Logger:      logger,
		PostRepo:    postRepo,
		NotiService: notificationService,
	}))

	// Route Group
	aiRoutes := router.Group("/ai")
	aiRoutes.Use(authMiddleware)
	{
		aiRoutes.POST("/:post_id/on", handler.OpenAIMode)
		aiRoutes.POST("/:post_id/off", handler.DisableOpenAIMode)
		aiRoutes.POST("/:post_id/chat", handler.Chat)
	}
}
