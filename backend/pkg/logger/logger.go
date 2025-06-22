package logger

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log *zap.Logger
)

// InitLogger sets up a global zap.Logger instance
func InitLogger(env string) {
	//  Ensure logs folder exists
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		if err := os.Mkdir("logs", 0755); err != nil {
			log.Fatalf("failed to create logs folder: %v", err)
		}
	}

	var cfg zap.Config

	if env == "production" {
		cfg = zap.NewProductionConfig()
		cfg.OutputPaths = []string{"stdout", "logs/logs.txt"}
		cfg.ErrorOutputPaths = []string{"stderr"}
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // "INFO", "ERROR" (no color)
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.OutputPaths = []string{"stdout", "logs/logs.txt"}
		cfg.ErrorOutputPaths = []string{"stderr"}
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

		cfg.EncoderConfig = zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder, // With color in dev
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
	}

	var err error
	Log, err = cfg.Build()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	zap.ReplaceGlobals(Log)
}

func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		Log.Info("HTTP Request",
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", duration),
		)
	}
}
