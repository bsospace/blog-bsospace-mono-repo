package main

import (
	"fmt"
	"log"
	"rag-searchbot-backend/api/v1/auth"
	"rag-searchbot-backend/api/v1/media"
	"rag-searchbot-backend/api/v1/post"
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/handlers"
	"rag-searchbot-backend/internal/cache"
	mediaInternal "rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/pkg/logger"
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
		log.Println("[Startup] ลบรูปภาพที่ไม่ได้ใช้งานทันทีตอนเริ่มเซิร์ฟเวอร์...")
		err := service.DeleteUnusedImages()
		if err != nil {
			log.Println("[Startup] ลบรูปภาพล้มเหลว:", err)
		} else {
			log.Println("[Startup] ลบรูปภาพที่ไม่ได้ใช้งานสำเร็จ")
		}
	}()

	// ตั้ง Cron ให้ลบทุกเที่ยงคืน
	_, err := c.AddFunc("0 0 0 * * *", func() {
		log.Println("[Cron] เริ่มลบรูปภาพที่ไม่ได้ใช้งาน...")
		err := service.DeleteUnusedImages()
		if err != nil {
			log.Println("[Cron] ลบรูปภาพล้มเหลว:", err)
		} else {
			log.Println("[Cron] ลบรูปภาพที่ไม่ได้ใช้งานสำเร็จ")
		}
	})

	if err != nil {
		log.Fatalln("ไม่สามารถตั้ง Cron Job ได้:", err)
	}

	c.Start()
}

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

	StartMediaCleanupCron(db, cacheService)

	r := gin.Default()

	var coreUrl string

	if cfg.AppEnv == "production" {
		coreUrl = cfg.CoreUrl
	} else {
		coreUrl = "http://bobby.posyayee.com:3000"
	}

	// CORS settings
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{coreUrl},
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

	r.POST("/upload", handlers.UploadHandler)
	r.POST("/ask", handlers.AskHandler)

	r.Run(":8088")
}
