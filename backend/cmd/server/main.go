package main

import (
	"fmt"
	"log"
	"os"
	"rag-searchbot-backend/api/v1/auth"
	"rag-searchbot-backend/api/v1/media"
	"rag-searchbot-backend/api/v1/post"
	"rag-searchbot-backend/api/v1/user"
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/handlers"
	"rag-searchbot-backend/internal/cache"
	mediaInternal "rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/pkg/logger"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Cron expression format explanation:
// "0 0 0 * * *"
//  ^ ^ ^ ^ ^ ^
//  | | | | | +--- Day of Week (0-6 or SUN-SAT)
//  | | | | +----- Month (1-12)
//  | | | +------- Day of Month (1-31)
//  | | +--------- Hour (0-23)
//  | +----------- Minute (0-59)
//  +------------- Second (0-59)

func StartMediaCleanupCron(db *gorm.DB, cache *cache.Service) {
	repo := mediaInternal.NewMediaRepository(db)
	service := mediaInternal.NewMediaService(repo)

	c := cron.New(cron.WithSeconds())

	// เรียกตอนเริ่ม server ทันที
	go func() {
		log.Println("[Startup] Deleting unused images...")
		err := service.DeleteUnusedImages()
		if err != nil {
			log.Println("[Startup] Fail to deleting image", err)
		} else {
			log.Println("[Startup] Deleted unused images successfully")
		}
	}()

	// ตั้ง Cron ให้ลบทุกเที่ยงคืน
	_, err := c.AddFunc("0 0 0 * * *", func() {
		log.Println("[Cron] Starting to delete unused images...")
		err := service.DeleteUnusedImages()
		if err != nil {
			log.Println("[Cron] Fail to deleting image: ", err)
		} else {
			log.Println("[Cron] Deleted unused images successfully")
		}
	})

	if err != nil {
		log.Fatalln("Can't start CRON :", err)
	}

	c.Start()
}

func main() {

	cfg := config.LoadConfig()

	logger.InitLogger(cfg.AppEnv)
	defer logger.Log.Sync()

	logger.Log.Info("Application started")

	// กำหนด Mode การทำงาน
	if cfg.AppEnv == "release" {
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
	cacheService := &cache.Service{
		Cache:       make(map[string]interface{}),
		RedisClient: redisClient,
		RedisTTL:    15 * time.Minute,
	}

	fmt.Println("=== Cache service initialized ===", cacheService)

	StartMediaCleanupCron(db, cacheService)

	r := gin.Default()

	var coreUrl []string

	if cfg.AppEnv == "release" || cfg.AppEnv == "production" {
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
	auth.RegisterRoutes(apiGroup, db, cacheService)
	post.RegisterRoutes(apiGroup, db, cacheService)
	media.RegisterRoutes(apiGroup, db, cacheService)
	user.RegisterRoutes(apiGroup, db, cacheService)

	r.POST("/upload", handlers.UploadHandler)
	r.POST("/ask", handlers.AskHandler)

	r.Run(":8088")
}
