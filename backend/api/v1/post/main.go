package post

import (
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/notification"
	"rag-searchbot-backend/internal/post"
	"rag-searchbot-backend/internal/queue"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/internal/ws"
	"rag-searchbot-backend/pkg/crypto"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, cache *cache.Service, logger *zap.Logger, asynqClient *asynq.Client, mux *asynq.ServeMux, socketManager *ws.Manager) {
	// Inject dependencies

	mediaRepo := media.NewMediaRepository(db)
	mediaService := media.NewMediaService(mediaRepo, logger)

	var postRepo post.PostRepositoryInterface = post.NewPostRepository(db)

	QueueRepository := queue.NewRepository(db)
	taskEnqueuer := post.NewTaskEnqueuer(asynqClient, QueueRepository)
	service := post.NewPostService(postRepo, mediaService, taskEnqueuer)

	handler := NewPostHandler(service)

	// สร้าง Repository ที่ใช้ GORM
	userRepository := user.NewRepository(db)

	// สร้าง Service ที่ใช้ Crypto
	crypto := crypto.NewCryptoService()

	// cryptoService
	cryptoService := crypto
	// สร้าง Service ที่ใช้ Repository
	userService := user.NewService(userRepository, cache)

	// notification service
	notificationRepo := notification.NewRepository(db)                              // Assuming you have a WebSocket manager
	notificationService := notification.NewService(notificationRepo, socketManager) // Assuming you have a WebSocket manager

	// Cache Service
	cacheService := cache

	// auth middleware
	authMiddleware := middleware.AuthMiddleware(userService, cryptoService, cacheService, logger)

	// EnqueueFilterPostContentByAI
	worker := post.FilterPostWorker{
		Logger:      logger,
		PostRepo:    postRepo,
		QueueRepo:   QueueRepository,
		NotiService: notificationService,
	}

	mux.HandleFunc(post.TaskTypeFilterPostContentByAI, post.FilterPostContentByAIWorkerHandler(worker))

	// Protected Category Routes
	postsRoutes := router.Group("/posts")
	// Public routes (no authentication required)
	{
		postsRoutes.GET("", handler.GetAll)
		postsRoutes.GET("/public/:username/:slug", handler.GetPublicPostBySlugAndUsername)
	}

	postsRoutes.Use(authMiddleware)
	{
		postsRoutes.POST("", handler.Create)
		postsRoutes.GET("/:short_slug", handler.GetByShortSlug)
		postsRoutes.GET("/my-posts", handler.MyPost)
		postsRoutes.PUT("/publish/:short_slug", handler.Publish)
		postsRoutes.PUT("/unpublish/:short_slug", handler.Unpublish)
		postsRoutes.DELETE("/:id", handler.Delete)
	}
}
