package ai

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

func NewEmbedPostWorkerHandler(logger *zap.Logger) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload EmbedPostPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			logger.Error("Failed to unmarshal task payload", zap.Error(err), zap.String("task_type", t.Type()))
			return err
		}

		postID := payload.Post.ID.String()
		logger.Info("Starting to embed post", zap.String("post_id", postID), zap.String("user_email", payload.User.Email))

		for i := 5; i > 0; i-- {
			logger.Info("Embedding post", zap.String("post_id", postID), zap.Int("attempt", 6-i))
			time.Sleep(1 * time.Second)
		}

		logger.Info("Post embedding completed", zap.String("post_id", postID), zap.String("user_email", payload.User.Email))
		return nil
	}
}
