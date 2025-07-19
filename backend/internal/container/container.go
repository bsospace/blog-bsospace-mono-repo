package container

import (
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/internal/ai"
	"rag-searchbot-backend/internal/auth"
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/internal/notification"
	"rag-searchbot-backend/internal/post"
	"rag-searchbot-backend/internal/queue"
	"rag-searchbot-backend/internal/user"
	"rag-searchbot-backend/internal/ws"
	"rag-searchbot-backend/pkg/crypto"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Container struct {
	Env *config.Config
	DB  *gorm.DB
	Log *zap.Logger

	UserService         user.ServiceInterface
	PostService         post.PostServiceInterface
	NotificationService notification.NotificationServiceInterface
	MediaService        media.MediaServiceInterface

	PostRepo  post.PostRepositoryInterface
	UserRepo  user.RepositoryInterface
	NotiRepo  notification.RepositoryInterface
	MediaRepo media.MediaRepositoryInterface

	// Add these fields for full DI
	CacheService                 cache.ServiceInterface
	SocketManager                *ws.Manager
	QueueRepo                    queue.QueueRepositoryInterface
	AsynqClient                  *asynq.Client
	AsynqMux                     *asynq.ServeMux
	CryptoService                *crypto.CryptoService
	AuthService                  auth.AuthServiceInterface
	AgentIntentClassifierService ai.AgentIntentClassifierServiceInterface
}

func NewContainer(
	env *config.Config,
	db *gorm.DB,
	log *zap.Logger,
	userRepo user.RepositoryInterface,
	postRepo post.PostRepositoryInterface,
	notiRepo notification.RepositoryInterface,
	mediaRepo media.MediaRepositoryInterface,
	userService user.ServiceInterface,
	postService post.PostServiceInterface,
	notiService notification.NotificationServiceInterface,
	mediaService media.MediaServiceInterface,
	cacheService cache.ServiceInterface,
	socketManager *ws.Manager,
	queueRepo queue.QueueRepositoryInterface,
	asynqClient *asynq.Client,
	asynqMux *asynq.ServeMux,
	cryptoService *crypto.CryptoService,
	authService auth.AuthServiceInterface,
	agentIntentClassifierService ai.AgentIntentClassifierServiceInterface,

) *Container {
	return &Container{
		Env:                          env,
		DB:                           db,
		Log:                          log,
		UserRepo:                     userRepo,
		PostRepo:                     postRepo,
		NotiRepo:                     notiRepo,
		MediaRepo:                    mediaRepo,
		UserService:                  userService,
		PostService:                  postService,
		NotificationService:          notiService,
		MediaService:                 mediaService,
		CacheService:                 cacheService,
		SocketManager:                socketManager,
		QueueRepo:                    queueRepo,
		AsynqClient:                  asynqClient,
		AsynqMux:                     asynqMux,
		CryptoService:                cryptoService,
		AuthService:                  authService,
		AgentIntentClassifierService: agentIntentClassifierService,
	}
}
