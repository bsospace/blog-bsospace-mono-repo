package post

import (
	"log"
	"rag-searchbot-backend/internal/container"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/notification"
	"rag-searchbot-backend/internal/post"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

func RegisterRoutes(router *gin.RouterGroup, container *container.Container, mux *asynq.ServeMux) {
	if container.AsynqClient == nil {
		log.Fatal("[FATAL] AsynqClient is nil! Cannot enqueue tasks.")
	}

	// Auth Middleware
	authMiddleware := middleware.NewAuthMiddleware(
		container.UserService,
		container.CryptoService,
		container.CacheService,
		container.Log,
	)

	// Worker Setup
	worker := post.FilterPostWorker{
		Logger:      container.Log,
		PostRepo:    container.PostRepo,
		QueueRepo:   container.QueueRepo,
		NotiService: container.NotificationService.(*notification.NotificationService),
	}

	// Create Service + Handler
	taskEnqueuer := post.NewTaskEnqueuer(container.AsynqClient, container.QueueRepo)
	postService := post.NewPostService(container.PostRepo, container.MediaService, taskEnqueuer)

	mux.HandleFunc(post.TaskTypeFilterPostContentByAI, post.FilterPostContentByAIWorkerHandler(worker))

	ps, ok := postService.(*post.PostService)
	if !ok {
		log.Fatal("[FATAL] Failed to cast postService to *post.PostService")
	}
	handler := NewPostHandler(ps)

	// Route Grouping
	postsRoutes := router.Group("/posts")

	// Public routes
	postsRoutes.GET("", handler.GetAll)
	postsRoutes.GET("/public/:username/:slug", handler.GetPublicPostBySlugAndUsername)

	// Protected routes
	postsRoutes.Use(authMiddleware.Handler())
	{
		postsRoutes.POST("", handler.Create)
		postsRoutes.GET("/:short_slug", handler.GetByShortSlug)
		postsRoutes.GET("/my-posts", handler.MyPost)
		postsRoutes.PUT("/publish/:short_slug", handler.Publish)
		postsRoutes.PUT("/unpublish/:short_slug", handler.Unpublish)
		postsRoutes.DELETE("/:id", handler.Delete)
	}
}
