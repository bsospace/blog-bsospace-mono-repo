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

const (
	ChunkSize   = 100 // Number of words per chunk
	OverlapSize = 10  // Number of overlapping words between chunks
)

func NewEmbedPostWorkerHandler(deps EmbedPostWorker) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload EmbedPostPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			deps.Logger.Error("Failed to unmarshal task payload", zap.Error(err), zap.String("task_type", t.Type()))
			return err
		}

		postID := payload.Post.ID.String()
		deps.Logger.Info("Starting to embed post", zap.String("post_id", postID), zap.String("user_email", payload.User.Email))

		existingPost, err := deps.PostRepo.GetByID(postID)
		if err != nil {
			deps.Logger.Error("Post not found", zap.Error(err), zap.String("post_id", postID))
			return err
		}
		deps.Logger.Info("Found post title", zap.String("title", existingPost.Title))

		if existingPost.Content == "" {
			err := errors.New("post has no HTML content")
			deps.Logger.Error("Empty content", zap.Error(err), zap.String("post_id", postID))
			return err
		}

		// Convert TipTap JSON content to plain text
		plainText := tiptap.ExtractTextFromTiptap(existingPost.Content)
		chunks := SplitTextToChunks(plainText, ChunkSize, OverlapSize)
		if len(chunks) == 0 {
			return errors.New("no chunks generated from content")
		}

		// Delete existing embeddings
		if err := deps.PostRepo.DeleteEmbeddingsByPostID(postID); err != nil {
			deps.Logger.Error("Failed to delete old embeddings", zap.Error(err), zap.String("post_id", postID))
			return err
		}

		// Generate and collect embeddings per chunk
		var embeddings []models.Embedding
		for _, chunk := range chunks {
			// log current chunk being processed
			deps.Logger.Debug("Processing chunk for embedding", zap.String("chunk", chunk))
			vec, err := GetEmbedding(chunk)
			if err != nil {
				deps.Logger.Error("Failed to get embedding for chunk", zap.Error(err), zap.String("chunk", chunk))
				return err
			}
			embeddings = append(embeddings, models.Embedding{
				PostID:  payload.Post.ID,
				Content: chunk,
				Vector:  pgvector.NewVector(vec),
			})
		}

		// Bulk insert new embeddings
		if err := deps.PostRepo.BulkInsertEmbeddings(&payload.Post, embeddings); err != nil {
			deps.Logger.Error("Failed to insert new embeddings", zap.Error(err), zap.String("post_id", postID))
			return err
		}
		deps.Logger.Info("Inserted all chunk embeddings", zap.String("post_id", postID), zap.Int("chunks", len(embeddings)))

		// Update post status
		updatedPost := payload.Post
		updatedPost.AIChatOpen = true
		updatedPost.AIReady = true
		updatedPost.Published = true
		if err := deps.PostRepo.Update(&updatedPost); err != nil {
			deps.Logger.Error("Failed to update post AIChatOpen", zap.Error(err), zap.String("post_id", postID))
			return err
		}

		// Send notification to user
		notiEvent := "notification:ai_mode_enabled"
		if err := deps.NotiService.Notify(
			&payload.User,
			fmt.Sprintf("Your post '%s' has been successfully enabled AI mode.", existingPost.Title),
			notiEvent,
			postID,
			nil,
		); err != nil {
			deps.Logger.Error("Failed to send notification", zap.Error(err), zap.String("post_id", postID))
			return err
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

func SplitTextToChunks(text string, chunkSize, overlap int) []string {
	words := strings.Fields(text)
	var chunks []string
	for i := 0; i < len(words); i += (chunkSize - overlap) {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, chunk)
		if end == len(words) {
			break
		}
	}
	return chunks
}
