package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/notification"
	"rag-searchbot-backend/internal/post"
	"strings"

	"rag-searchbot-backend/pkg/tiptap"

	"github.com/hibiken/asynq"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

type EmbedPostWorker struct {
	Logger      *zap.Logger
	PostRepo    post.PostRepositoryInterface
	NotiService *notification.NotificationService
}

func NewEmbedPostWorkerHandler(deps EmbedPostWorker) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload EmbedPostPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			deps.Logger.Error("Failed to unmarshal task payload", zap.Error(err), zap.String("task_type", t.Type()))
			return err
		}

		postID := payload.Post.ID.String()
		deps.Logger.Info("Starting to embed post", zap.String("post_id", postID), zap.String("user_email", payload.User.Email))

		// Optional: do something with PostRepo
		existingPost, err := deps.PostRepo.GetByID(postID)
		if err != nil {
			deps.Logger.Error("Post not found", zap.Error(err), zap.String("post_id", postID))
			return err
		}
		deps.Logger.Info("Found post title", zap.String("title", existingPost.Title))

		if existingPost.Content == "" {
			err := errors.New("post has no HTML content")
			deps.Logger.Error("Failed to get embedding", zap.Error(err), zap.String("post_id", postID))
			return err
		}

		// post content to plain text embedding model
		plainText := tiptap.ExtractTextFromTiptap(existingPost.Content)
		embedding, err := GetEmbedding(plainText)
		if err != nil {
			deps.Logger.Error("Failed to get embedding", zap.Error(err), zap.String("post_id", postID))
			return err
		}

		// get existin embedding
		existingEmbedding, err := deps.PostRepo.GetEmbeddingByPostID(payload.Post.ID.String())

		if err != nil {
			deps.Logger.Error("Failed to get existing embedding", zap.Error(err), zap.String("post_id", postID))
			return err
		}

		embeddingModel := models.Embedding{
			Vector: pgvector.NewVector(embedding),
		}

		if len(existingEmbedding) > 0 {
			// Update existing embedding
			embeddingModel.ID = existingEmbedding[0].ID
			embeddingModel.PostID = existingEmbedding[0].PostID
			if err := deps.PostRepo.UpdateEmbedding(&payload.Post, embeddingModel); err != nil {
				deps.Logger.Error("Failed to update embedding", zap.Error(err), zap.String("post_id", postID))
				return err
			}
			deps.Logger.Info("Updated existing embedding", zap.String("post_id", postID))

			// set AIChatOpen to true
			updatedPost := payload.Post
			updatedPost.AIChatOpen = true
			updatedPost.AIReady = true
			if err := deps.PostRepo.Update(&updatedPost); err != nil {
				deps.Logger.Error("Failed to update post AIChatOpen", zap.Error(err), zap.String("post_id", postID))
				return err
			}

			// send notification to user
			notiEvent := "notification:" + payload.Post.Title + ":ai_mode_enabled"
			if err := deps.NotiService.Notify(
				&payload.User,
				"AI Mode Enabled",
				notiEvent,
				fmt.Sprintf("Your post '%s' has been successfully eanbled AI mode.", existingPost.Title),
				nil,
			); err != nil {
				deps.Logger.Error("Failed to send notification", zap.Error(err), zap.String("post_id", postID))
				return err
			}

		} else {
			// Insert new embedding
			embeddingModel.PostID = payload.Post.ID
			if err := deps.PostRepo.InsertEmbedding(&payload.Post, embeddingModel); err != nil {
				deps.Logger.Error("Failed to insert embedding", zap.Error(err), zap.String("post_id", postID))
				return err
			}
			deps.Logger.Info("Inserted new embedding", zap.String("post_id", postID))

			// set AIChatOpen to true
			updatedPost := payload.Post
			updatedPost.AIChatOpen = true
			updatedPost.AIReady = true
			if err := deps.PostRepo.Update(&updatedPost); err != nil {
				deps.Logger.Error("Failed to update post AIChatOpen", zap.Error(err), zap.String("post_id", postID))
				return err
			}

			// send notification to user
			notiEvent := "notification:" + payload.Post.Title + ":ai_mode_enabled"
			if err := deps.NotiService.Notify(
				&payload.User,
				"AI Mode Enabled",
				notiEvent,
				fmt.Sprintf("Your post '%s' has been successfully enabled AI mode.", existingPost.Title),
				nil,
			); err != nil {
				deps.Logger.Error("Failed to send notification", zap.Error(err), zap.String("post_id", postID))
				return err
			}
		}

		deps.Logger.Info("Post embedding completed", zap.String("post_id", postID), zap.String("user_email", payload.User.Email))
		return nil
	}
}

func FilterPostContentByAIWorkerHandler(deps EmbedPostWorker) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload FilterPostContentByAIPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			deps.Logger.Error("Failed to parse payload", zap.Error(err))
			return err
		}

		deps.Logger.Info("Filtering post content", zap.String("post_id", payload.Post.ID.String()))

		// Skip filtering if content is too short
		if len(payload.Post.Content) < 100 {
			deps.Logger.Info("Content too short, skipping AI check", zap.String("post_id", payload.Post.ID.String()))
			return nil
		}

		prompt := fmt.Sprintf(`Please check if this blog content contains profanity, 18+ content, or spam.
If the content is safe, respond with "SAFE"
If not safe, respond with "UNSAFE" and provide the reason.

Content:
%s`, payload.Post.Content)

		body := OllamaRequest{
			Model:  os.Getenv("AI_MODEL"),
			Prompt: prompt,
			Stream: false,
		}

		data, err := json.Marshal(body)
		if err != nil {
			deps.Logger.Error("Failed to marshal Ollama request", zap.Error(err))
			return err
		}

		resp, err := http.Post(
			os.Getenv("OLLAMA_HOST")+"/api/generate",
			"application/json",
			bytes.NewBuffer(data),
		)
		if err != nil {
			deps.Logger.Error("Failed to call Ollama", zap.Error(err))
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("ollama responded with status: %d", resp.StatusCode)
			deps.Logger.Error("Ollama returned non-200", zap.Error(err))
			return err
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			deps.Logger.Error("Failed to read Ollama response", zap.Error(err))
			return err
		}

		var ollamaResp OllamaResponse
		if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
			deps.Logger.Error("Failed to decode Ollama response", zap.Error(err))
			return err
		}

		if ollamaResp.Response == "" {
			err = errors.New("ollama response is empty")
			deps.Logger.Error("Empty response from Ollama", zap.Error(err))
			return err
		}

		result := strings.TrimSpace(ollamaResp.Response)
		deps.Logger.Info("Ollama content filter result", zap.String("result", result))

		// Optional: ถ้าอยาก block content ไม่ปลอดภัย
		if strings.HasPrefix(result, "UNSAFE") {
			return errors.New("post flagged as UNSAFE: " + result)
		}

		return nil
	}
}
