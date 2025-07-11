//go:build wireinject
// +build wireinject

package container

import (
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/internal/auth"
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/internal/notification"
	"rag-searchbot-backend/internal/post"
	"rag-searchbot-backend/internal/queue"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/internal/ws"
	"rag-searchbot-backend/pkg/crypto"

	"time"

	"github.com/google/wire"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var postSet = wire.NewSet(
	post.NewPostRepository,
	post.NewTaskEnqueuer,
	post.NewPostService,
)

var userSet = wire.NewSet(
	user.NewRepository,
	user.NewService,
)

var mediaSet = wire.NewSet(
	media.NewMediaRepository,
	media.NewMediaService,
)

var notificationSet = wire.NewSet(
	notification.NewRepository,
	notification.NewService,
)

var authSet = wire.NewSet(
	auth.NewAuthService,
)

var CrypetoSet = wire.NewSet(
	crypto.NewCryptoService,
)

func NewCacheService(redisClient *redis.Client, redisTTL time.Duration) cache.ServiceInterface {
	return cache.NewService(redisClient, redisTTL)
}

func NewAsynqMux() *asynq.ServeMux {
	return asynq.NewServeMux()
}

func InitializeContainer(
	env *config.Config,
	db *gorm.DB,
	log *zap.Logger,
	redisClient *redis.Client,
	redisTTL time.Duration,
	asynqClient *asynq.Client,
	queueRepo queue.QueueRepositoryInterface,
) (*Container, error) {
	wire.Build(
		NewContainer,
		userSet,
		postSet,
		notificationSet,
		mediaSet,
		NewCacheService,
		ws.NewManager,
		CrypetoSet,
		NewAsynqMux,
		authSet,
	)
	return &Container{}, nil
}
