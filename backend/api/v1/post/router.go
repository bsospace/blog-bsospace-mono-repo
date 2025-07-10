package post

import (
	"rag-searchbot-backend/internal/container"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/internal/notification"
	"rag-searchbot-backend/internal/post"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, container *container.Container) {
	handler := NewPostHandler(container.PostService.(*post.PostService))
	authMiddleware := middleware.NewAuthMiddleware(
		container.UserService,
		container.CryptoService,
		container.CacheService,
		container.Log,
	)
	worker := post.FilterPostWorker{
		Logger:      container.Log,
		PostRepo:    container.PostRepo,
		QueueRepo:   container.QueueRepo,
		NotiService: container.NotificationService.(*notification.NotificationService),
	}
	container.AsynqMux.HandleFunc(post.TaskTypeFilterPostContentByAI, post.FilterPostContentByAIWorkerHandler(worker))
	postsRoutes := router.Group("/posts")
	postsRoutes.GET("", handler.GetAll)
	postsRoutes.GET("/public/:username/:slug", handler.GetPublicPostBySlugAndUsername)
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
