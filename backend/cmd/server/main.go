package main

import (
	"log"
	"os"
	"rag-searchbot-backend/api/v1/ai"
	"rag-searchbot-backend/api/v1/auth"
	"rag-searchbot-backend/api/v1/media"
	"rag-searchbot-backend/api/v1/notification"
	"rag-searchbot-backend/api/v1/post"
	"rag-searchbot-backend/api/v1/user"
	"rag-searchbot-backend/api/v1/ws"
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/internal/container"
	mediaInternal "rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/internal/middleware"
	"rag-searchbot-backend/pkg/logger"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Cron expression format explanation:
// "0 0 0 * * *"
//
//	^ ^ ^ ^ ^ ^
//	| | | | | +--- Day of Week (0-6 or SUN-SAT)
//	| | | | +----- Month (1-12)
//	| | | +------- Day of Month (1-31)
//	| | +--------- Hour (0-23)
//	| +----------- Minute (0-59)
func StartMediaCleanupCron(db *gorm.DB, cache *cache.Service, logger *zap.Logger) {
	repo := mediaInternal.NewMediaRepository(db)
	service := mediaInternal.NewMediaService(repo, logger)

	c := cron.New(cron.WithSeconds())

	// เรียกตอนเริ่ม server ทันที
	go func() {
		logger.Info("[Startup] Starting to delete unused images...")
		err := service.DeleteUnusedImages()
		if err != nil {
			logger.Error("[Startup] Fail to deleting image", zap.Error(err))
		} else {
			logger.Info("[Startup] Deleted unused images successfully")
		}
	}()

	// ตั้ง Cron ให้ลบทุกเที่ยงคืน
	_, err := c.AddFunc("0 0 0 * * *", func() {
		logger.Info("[Cron] Starting to delete unused images...")
		err := service.DeleteUnusedImages()
		if err != nil {
			logger.Error("[Cron] Failed to delete unused images", zap.Error(err))
		} else {
			logger.Info("[Cron] Deleted unused images successfully")
		}
	})

	if err != nil {
		logger.Error("[Cron] Failed to schedule media cleanup", zap.Error(err))
	} else {
		logger.Info("[Cron] Media cleanup scheduled to run daily at midnight")
	}

	c.Start()
}

// applyRouteRateLimiting applies rate limiting to specific routes based on configuration
func applyRouteRateLimiting(router *gin.RouterGroup, settings *config.RateLimitSettings, redisClient *redis.Client, logger *zap.Logger) {
	for route, config := range settings.Routes {
		if !config.Enabled {
			continue
		}

		rateLimitConfig := middleware.RateLimitConfig{
			Strategy:       config.Strategy,
			MaxRequests:    config.MaxRequests,
			WindowSize:     config.WindowSize,
			RefillRate:     config.RefillRate,
			BucketCapacity: config.BucketCapacity,
			UseRedis:       settings.Redis.Enabled && redisClient != nil,
			RedisClient:    redisClient,
			RedisKeyPrefix: settings.Redis.KeyPrefix + ":route:" + route,
			Logger:         logger,
		}

		// Apply rate limiting to the specific route
		router.Use(func(c *gin.Context) {
			if c.Request.URL.Path == route {
				middleware.RateLimitMiddleware(rateLimitConfig)(c)
			}
		})

		logger.Info("Route rate limiting applied",
			zap.String("route", route),
			zap.String("strategy", config.Strategy),
			zap.Int("max_requests", config.MaxRequests),
			zap.Duration("window_size", config.WindowSize))
	}
}

func main() {

	cfg := config.LoadConfig()

	logger.InitLogger(cfg.AppEnv)
	defer logger.Log.Sync()

	logger.Log.Info("Application started")

	// กำหนด Mode การทำงาน
	if cfg.AppEnv == "release" {

		gin.SetMode(gin.ReleaseMode)
		logger.Log.Info("Running in Production Mode")
	} else {
		gin.SetMode(gin.DebugMode)
		logger.Log.Info("Running in Development Mode")
	}

	// เชื่อมต่อฐานข้อมูล
	db := config.ConnectDatabase()

	if db == nil {
		log.Fatal("Failed to connect to database")
	} else {
		logger.Log.Info("Database connection established successfully")
	}

	redisClient := config.ConnectRedis()

	if redisClient == nil {
		logger.Log.Fatal("Failed to connect to Redis")
	} else {
		logger.Log.Info("Redis connection established successfully")
	}

	// Load rate limiting configuration
	rateLimitSettings := config.LoadRateLimitSettings()
	logger.Log.Info("Rate limiting configuration loaded",
		zap.Bool("global_enabled", rateLimitSettings.Global.Enabled),
		zap.Bool("ip_enabled", rateLimitSettings.IP.Enabled),
		zap.Bool("api_enabled", rateLimitSettings.API.Enabled))

	// TTL 15 minutes
	cacheService := &cache.Service{
		Cache:       make(map[string]interface{}),
		RedisClient: redisClient,
		RedisTTL:    24 * time.Hour,
	}

	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr: cfg.RedisAddr,
	})

	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{Concurrency: 10},
	)

	mux := asynq.NewServeMux()

	go func() {
		if err := asynqServer.Run(mux); err != nil {
			logger.Log.Fatal("Worker error", zap.Error(err))
		}
	}()

	logger.Log.Info("Cache service initialized successfully")

	StartMediaCleanupCron(db, cacheService, logger.Log)

	containerDI, err := container.InitializeContainer(&cfg, db, logger.Log, redisClient, 24*time.Hour, asynqClient)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.Use(logger.ZapLogger())
	r.Use(gin.Recovery())

	// Apply global rate limiting if enabled
	if rateLimitSettings.Global.Enabled {
		globalRateLimit := middleware.RateLimitMiddleware(middleware.RateLimitConfig{
			Strategy:       rateLimitSettings.Global.Strategy,
			MaxRequests:    rateLimitSettings.Global.MaxRequests,
			WindowSize:     rateLimitSettings.Global.WindowSize,
			UseRedis:       rateLimitSettings.Redis.Enabled && redisClient != nil,
			RedisClient:    redisClient,
			RedisKeyPrefix: rateLimitSettings.Redis.KeyPrefix + ":global",
			Logger:         logger.Log,
		})
		r.Use(globalRateLimit)
		logger.Log.Info("Global rate limiting applied",
			zap.String("strategy", rateLimitSettings.Global.Strategy),
			zap.Int("max_requests", rateLimitSettings.Global.MaxRequests),
			zap.Duration("window_size", rateLimitSettings.Global.WindowSize))
	}

	var coreUrl []string

	if cfg.AppEnv == "release" {
		coreUrl = strings.Split(os.Getenv("ALLOWED_ORIGINS_PROD"), ",")
	} else {
		coreUrl = strings.Split(os.Getenv("ALLOWED_ORIGINS_DEV"), ",")
	}

	// CORS settings
	r.Use(cors.New(cors.Config{
		AllowOrigins:     coreUrl,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/api/v1", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Rag Search Bot API",
			"status":  "ok",
		})
	})

	apiGroup := r.Group("/api/v1")

	// Apply route-specific rate limiting
	applyRouteRateLimiting(apiGroup, rateLimitSettings, redisClient, logger.Log)

	ws.StartWebSocketServer(apiGroup, containerDI)
	auth.RegisterRoutes(apiGroup, containerDI)
	post.RegisterRoutes(apiGroup, containerDI, mux)
	media.RegisterRoutes(apiGroup, containerDI)
	user.RegisterRoutes(apiGroup, containerDI)
	ai.RegisterRoutes(apiGroup, containerDI, mux)
	notification.RegisterRoutes(apiGroup, containerDI)

	r.Run(":8088")
}
