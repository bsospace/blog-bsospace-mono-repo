package post

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
	"rag-searchbot-backend/internal/queue"
	"rag-searchbot-backend/pkg/logger"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type FilterPostWorker struct {
	Logger    *zap.Logger
	PostRepo  PostRepositoryInterface
	QueueRepo queue.QueueRepositoryInterface
}

func FilterPostContentByAIWorkerHandler(deps FilterPostWorker) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload FilterPostContentByAIPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			deps.Logger.Error("Failed to parse payload", zap.Error(err))
			return err
		}

		startedAt := time.Now()
		deps.Logger.Info("Filtering post content", zap.String("post_id", payload.Post.ID.String()), zap.String("post_title", payload.Post.Title))

		if len(payload.Post.Content) < 100 {
			return handleSkippedContent(deps, t, &payload, startedAt)
		}

		prompt := buildModerationPrompt(payload)
		result, err := getModerationResult(prompt)
		if err != nil {
			deps.Logger.Error("AI moderation failed", zap.Error(err))
			return err
		}

		deps.Logger.Info("Content filter result", zap.String("result", result))

		status := "SUCCESS"
		message := "Content filtered successfully"

		if strings.HasPrefix(result, "UNSAFE") {
			status = "UNSAFE"
			moderationResult := parseModerationResult(result)
			if moderationResult == nil {
				return fmt.Errorf("failed to parse moderation result: %s", result)
			}

			// format the message with reason and content type
			message = fmt.Sprintf("Your post may contain inappropriate content: %s\nDetected category: %s", moderationResult.Reason, moderationResult.ContentType)
			deps.Logger.Warn("Post content flagged as UNSAFE", zap.String("reason", moderationResult.Reason), zap.String("content_type", moderationResult.ContentType))
		}

		if err := updateTaskLog(deps, t, &payload, startedAt, status, message); err != nil {
			deps.Logger.Error("Failed to update task log", zap.Error(err))
			return err
		}

		if status == "UNSAFE" {
			deps.Logger.Warn("Content blocked", zap.String("post_id", payload.Post.ID.String()))
			// update post status to rejected
			if err := UpdatePublishPostResult(deps, payload.Post.ID.String(), status, message); err != nil {
				deps.Logger.Error("Failed to update post status", zap.String("post_id", payload.Post.ID.String()), zap.Error(err))
				return err
			}
			return nil
		}

		// Update post status based on moderation result
		if err := UpdatePublishPostResult(deps, payload.Post.ID.String(), status, message); err != nil {
			deps.Logger.Error("Failed to update post status", zap.String("post_id", payload.Post.ID.String()), zap.Error(err))
			return err
		}

		return nil
	}
}

func handleSkippedContent(deps FilterPostWorker, t *asynq.Task, payload *FilterPostContentByAIPayload, startedAt time.Time) error {
	taskLog := &models.QueueTaskLog{
		TaskID:     t.ResultWriter().TaskID(),
		TaskType:   TaskTypeFilterPostContentByAI,
		RefID:      payload.Post.ID.String(),
		RefType:    "POST",
		Status:     "SKIPPED",
		Message:    "This is spam or too short content, skipping AI check and can't be published",
		StartedAt:  startedAt,
		FinishedAt: time.Now(),
		Duration:   int64(time.Since(startedAt) / time.Millisecond),
		Payload:    string(t.Payload()),
		UserID:     payload.User.ID,
	}
	if err := deps.QueueRepo.UpdateStatusByTask(taskLog); err != nil {
		deps.Logger.Error("Failed to create task log for skipped content", zap.Error(err))
		return err
	}
	deps.Logger.Info("Post content skipped due to short length", zap.String("post_id", payload.Post.ID.String()), zap.String("post_title", payload.Post.Title))

	// Update post status to rejected
	if err := UpdatePublishPostResult(deps, payload.Post.ID.String(), "UNSAFE", "This is spam or too short content, skipping AI check and can't be published"); err != nil {
		deps.Logger.Error("Failed to update post status for skipped content", zap.String("post_id", payload.Post.ID.String()), zap.Error(err))
		return err
	}
	return nil
}

type ModerationResult struct {
	Status      string // SAFE or UNSAFE
	Reason      string // Optional (only for UNSAFE)
	ContentType string // Always present
}

func parseModerationResult(response string) *ModerationResult {
	response = strings.TrimSpace(response)

	if strings.HasPrefix(response, "SAFE:") {
		contentType := strings.TrimSpace(strings.TrimPrefix(response, "SAFE:"))
		return &ModerationResult{
			Status:      "SAFE",
			Reason:      "",
			ContentType: contentType,
		}
	}

	if strings.HasPrefix(response, "UNSAFE:") {
		rest := strings.TrimSpace(strings.TrimPrefix(response, "UNSAFE:"))
		parts := strings.SplitN(rest, "|", 2)
		if len(parts) != 2 {
			return nil
		}
		reason := strings.TrimSpace(parts[0])
		contentType := strings.TrimSpace(parts[1])
		return &ModerationResult{
			Status:      "UNSAFE",
			Reason:      reason,
			ContentType: contentType,
		}
	}

	return nil
}
func buildModerationPrompt(payload FilterPostContentByAIPayload) string {
	return fmt.Sprintf(`You are a strict AI content moderator.

Carefully review the content and respond with one of the following formats ONLY:
- SAFE: [REASON] | [CONTENT_TYPE]
- UNSAFE: [REASON including specific words or phrases] | [CONTENT_TYPE]

CONTENT_TYPE must be one of:
ARTICLE, NEWS, TUTORIAL, REVIEW, ADVERTISEMENT, ENTERTAINMENT, ADULT, SPAM, SEXUAL, VIOLENT, HATE, PROVOCATIVE, MISINFORMATION, POLITICAL, HEALTH, FINANCIAL, EDUCATIONAL, OTHER.

Guidelines:
- If the content includes profanity, sexual language, vulgar jokes, violent speech, hate speech, spam, or adult references — mark as UNSAFE and include the word/phrase.
- If the content is extremely short, meaningless, repetitive, or test-like — mark as UNSAFE with reason "Meaningless or test content".
- If the content has no clear meaning, is just a few words, or lacks structure — mark as UNSAFE with reason "Lacks informative or meaningful value".
- If the content is structured and informative (like articles, guides, tutorials, opinions), mark as SAFE with brief reason and type.
- If content is just one or two meaningless or ambiguous words — mark as UNSAFE: Lacks meaning | OTHER.
- Check for Thai slang words like: กาก, มึง, กู, เหี้ย, ควย, สัส, ดอก, จิ๋ม หี etc.
- Check for English slang/rude words like: wtf, stfu, fck, shit, damn, ass etc.

Examples:
- SAFE: Clear tutorial content | TUTORIAL
- UNSAFE: Contains profanity 'f***' | SEXUAL 
- UNSAFE: Contains Thai slang 'มึง' | HATE
- UNSAFE: Just one meaningless word 'ดี้จา' | OTHER
- UNSAFE: Test content '....' | OTHER
- SAFE: Informative article about health benefits of exercise | HEALTH

--- Content to Review ---

Title: %s
Description: %s
Content: %s

--- End ---`, payload.Post.Title, payload.Post.Description, payload.Post.Content)
}

func getModerationResult(prompt string) (string, error) {
	if os.Getenv("AI_SELF_HOST") == "true" {
		return callOllama(prompt)
	}
	return callOpenRouter(prompt)
}

func callOllama(prompt string) (string, error) {
	body := OllamaRequest{
		Model:  os.Getenv("AI_MODEL"),
		Prompt: prompt,
		Stream: false,
	}
	data, _ := json.Marshal(body)
	resp, err := http.Post(os.Getenv("AI_HOST")+"/api/generate", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama responded with status: %d", resp.StatusCode)
	}
	var ollamaResp OllamaResponse
	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, &ollamaResp)
	return strings.TrimSpace(ollamaResp.Response), nil
}

func callOpenRouter(prompt string) (string, error) {
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		return "", errors.New("AI_API_KEY not set for OpenRouter")
	}
	openRouterPayload := map[string]interface{}{
		"model": os.Getenv("AI_MODEL"),
		"messages": []map[string]string{
			{"role": "system", "content": "You are an AI content moderator."},
			{"role": "user", "content": prompt},
		},
	}
	reqBody, _ := json.Marshal(openRouterPayload)
	req, _ := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openrouter responded with status: %d", resp.StatusCode)
	}
	var parsed map[string]interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &parsed)
	choices := parsed["choices"].([]interface{})
	message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	return strings.TrimSpace(message["content"].(string)), nil
}

func updateTaskLog(deps FilterPostWorker, t *asynq.Task, payload *FilterPostContentByAIPayload, startedAt time.Time, status, message string) error {
	finishedAt := time.Now()
	taskLog := &models.QueueTaskLog{
		TaskID:     t.ResultWriter().TaskID(),
		TaskType:   TaskTypeFilterPostContentByAI,
		RefID:      payload.Post.ID.String(),
		RefType:    "POST",
		Status:     status,
		Message:    message,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		Duration:   int64(finishedAt.Sub(startedAt) / time.Millisecond),
		Payload:    string(t.Payload()),
		UserID:     payload.User.ID,
	}
	return deps.QueueRepo.UpdateStatusByTask(taskLog)
}

func UpdatePublishPostResult(
	deps FilterPostWorker,
	postID string,
	status string,
	message string,
) error {
	post, err := deps.PostRepo.GetByID(postID)
	if err != nil {
		deps.Logger.Error("Failed to get post by ID", zap.String("post_id", postID), zap.Error(err))
		return err
	}

	logger.Log.Info("Updating post status", zap.String("post_id", postID), zap.String("status", status), zap.String("message", message))

	switch status {
	case "SUCCESS":
		post.Status = models.PostPublished
	case "UNSAFE":
		post.Status = models.PostRejected
	default:
		post.Status = models.PostRejected
	}

	now := time.Now()
	post.PublishedAt = &now
	post.Published = (status == "SUCCESS")

	if err := deps.PostRepo.Update(post); err != nil {
		deps.Logger.Error("Failed to update post status", zap.String("post_id", postID), zap.Error(err))
		return err
	}
	deps.Logger.Info("Post status updated", zap.String("post_id", postID), zap.String("status", status), zap.String("message", message))
	return nil
}
