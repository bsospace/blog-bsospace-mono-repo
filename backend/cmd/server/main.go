package main

import (
	"fmt"
	"log"
	"rag-searchbot-backend/api/v1/auth"
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/handlers"
	"rag-searchbot-backend/internal/cache"
	"rag-searchbot-backend/pkg/logger"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	cfg := config.LoadConfig()

	logger.InitLogger(cfg.AppEnv)
	defer logger.Log.Sync()

	logger.Log.Info("Application started")

	// กำหนด Mode การทำงาน
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
		log.Println("Running in Production Mode")
	} else {
		gin.SetMode(gin.DebugMode)
		log.Println("Running in Development Mode")
	}

	// เชื่อมต่อฐานข้อมูล
	db := config.ConnectDatabase()

	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	redisClient := config.ConnectRedis()

	if redisClient == nil {
		log.Fatal("Failed to connect to Redis")
	}

	// TTL 15 minutes
	cacheService := cache.NewService(redisClient, 15*time.Minute)
	fmt.Println("=== Cache service initialized ===", cacheService)

	r := gin.Default()

	// CORS settings
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://192.168.1.105:5173"},
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
	auth.RegisterRoutes(apiGroup, db, cacheService)

	r.POST("/upload", handlers.UploadHandler)
	r.POST("/ask", handlers.AskHandler)

	r.Run(":8088")
}
