package ai

import (
	"rag-searchbot-backend/internal/ai"
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/post"
	"rag-searchbot-backend/internal/user"
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
) {
	// Repository
	postRepo := post.NewPostRepository(db)

	taskEnqueuer := ai.NewTaskEnqueuer(asynqClient)

	// AI Service
	aiService := ai.NewAIService(postRepo, taskEnqueuer)

	// Handler
	handler := NewAIHandler(aiService)

	// Middleware
	cryptoService := crypto.NewCryptoService()
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo, cache)
	authMiddleware := middleware.AuthMiddleware(userService, cryptoService, cache, logger)

	// Register AI worker
	// Register AI worker with full dependency injection
	mux.HandleFunc(ai.TaskTypeEmbedPost, ai.NewEmbedPostWorkerHandler(ai.EmbedPostWorker{
		Logger:   logger,
		PostRepo: postRepo,
	}))

	// Route Group
	aiRoutes := router.Group("/ai")
	aiRoutes.Use(authMiddleware)
	{
		aiRoutes.POST("/:post_id/on", handler.OpenAIMode)
	}
}
