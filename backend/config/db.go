package config

import (
	"context"
	"log"
	"rag-searchbot-backend/internal/models"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ตัวแปร DB ใช้เก็บ instance ของฐานข้อมูล
var DB *gorm.DB

// ConnectDatabase เชื่อมต่อฐานข้อมูล PostgreSQL และทำ Migration
func ConnectDatabase() *gorm.DB {
	// โหลดค่าจาก .env
	config := LoadConfig()

	// ใช้ DATABASE_URL จาก .env
	dsn := config.DatabaseURL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(map[bool]logger.LogLevel{true: logger.Silent, false: logger.Info}[config.AppEnv == "production"]),
	})

	if err != nil {
		log.Fatal("[ERROR] Failed to connect to database: ", err)
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		log.Fatalf("Failed to create extension vector: %v", err)
	}
	if err := db.Exec("ALTER TABLE ai_responses ALTER COLUMN embedding_id DROP NOT NULL").Error; err != nil {
		log.Fatalf("Failed to alter table ai_responses: %v", err)
	}

	// ทำ Migration ให้กับทุกตาราง
	err = db.Set("gorm:foreign_key_constraints", true).AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Comment{},
		&models.Tag{},
		&models.Category{},
		&models.Embedding{},
		&models.Notification{},
		&models.AIUsageLog{},
		&models.AIResponse{},
		&models.ImageUpload{},
		&models.QueueTaskLog{},
	)

	if err != nil {
		log.Fatal("[ERROR] Migration failed:", err)
	}

	// เก็บ instance ของ DB ไว้ในตัวแปร DB
	DB = db
	log.Println("[INFO] Database connected & migration completed successfully!")

	return DB
}

func ConnectRedis() *redis.Client {

	log.Println("[INFO] Connecting to Redis...")
	config := LoadConfig()
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisHost + ":" + "6379",
		Password: config.RedisPassword,
		DB:       0,
		Protocol: 2,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis: ", err)
	} else {
		log.Println("Redis connected successfully!")
	}
	return client
}
