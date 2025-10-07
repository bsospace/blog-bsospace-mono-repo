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
	"rag-searchbot-backend/internal/container"
	"rag-searchbot-backend/pkg/logger"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

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

	containerDI, err := container.InitializeContainer(&cfg, db, logger.Log, redisClient, 24*time.Hour, asynqClient)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.Use(logger.ZapLogger())
	r.Use(gin.Recovery())

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
	ws.StartWebSocketServer(apiGroup, containerDI)
	auth.RegisterRoutes(apiGroup, containerDI)
	post.RegisterRoutes(apiGroup, containerDI, mux)
	media.RegisterRoutes(apiGroup, containerDI)
	user.RegisterRoutes(apiGroup, containerDI)
	ai.RegisterRoutes(apiGroup, containerDI, mux)
	notification.RegisterRoutes(apiGroup, containerDI)

	r.Run(":8088")
}
