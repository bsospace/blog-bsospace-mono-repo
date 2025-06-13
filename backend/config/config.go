package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	Port             string
	CoreUrl          string
	AppEnv           string
	AppUrl           string
	RedisHost        string
	RedisPort        string
	RedisPassword    string
	OpenIDURL        string
	OllamaHost       string
	ChibisafeURL     string
	ChibisafeKey     string
	ChibisafeAlbumId string
}

func LoadConfig() Config {
	// Try loading `.env` from the project root or `server/v2/`
	paths := []string{".env", "../../.env"}

	var loaded bool
	for _, path := range paths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("Loaded .env from: %s\n", path)
			loaded = true
			break
		}
	}

	if !loaded {
		log.Println("Warning: No .env file found in expected locations")
	}

	// Read env variables with default fallback
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_EXTERNAL_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	return Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		CoreUrl:          os.Getenv("APP_CORE_URL"),
		Port:             os.Getenv("APP_PORT"),
		AppEnv:           os.Getenv("APP_ENV"),
		AppUrl:           os.Getenv("APP_URL"),
		OpenIDURL:        os.Getenv("OPEN_ID_URL"),
		OllamaHost:       os.Getenv("OLLAMA_HOST"),
		ChibisafeURL:     os.Getenv("CHIBISAFE_URL"),
		ChibisafeKey:     os.Getenv("CHIBISAFE_KEY"),
		ChibisafeAlbumId: os.Getenv("CHIBISAFE_ALBUM_ID"),
		RedisHost:        redisHost,
		RedisPort:        redisPort,
		RedisPassword:    redisPassword,
	}
}
