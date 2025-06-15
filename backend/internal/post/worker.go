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
	"rag-searchbot-backend/internal/notification"
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
	Logger      *zap.Logger
	PostRepo    PostRepositoryInterface
	QueueRepo   queue.QueueRepositoryInterface
	NotiService *notification.NotificationService
}

func FilterPostContentByAIWorkerHandler(deps FilterPostWorker) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload FilterPostContentByAIPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			deps.Logger.Error("Failed to parse payload", zap.Error(err))
			return err
		}

		startedAt := time.Now()
		deps.Logger.Info("Filtering post content",
			zap.String("post_id", payload.Post.ID.String()),
			zap.String("post_title", payload.Post.Title),
		)

		if len(*payload.Post.HTMLContent) < 100 {
			return handleSkippedContent(deps, t, &payload, startedAt)
		}

		prompt := buildModerationPrompt(payload)
		result, err := getModerationResult(prompt)
		if err != nil {
			deps.Logger.Error("AI moderation failed", zap.Error(err))
			return err
		}

		deps.Logger.Info("Content filter result", zap.String("result", result))

		// Default values
		status := "SUCCESS"
		message := "Content filtered successfully"

		if strings.HasPrefix(result, "UNSAFE") {
			status = "UNSAFE"
			moderation := parseModerationResult(result)
			if moderation == nil {
				return fmt.Errorf("failed to parse moderation result: %s", result)
			}

			message = fmt.Sprintf(
				"Your post may contain inappropriate content: %s\nDetected category: %s",
				moderation.Reason, moderation.ContentType,
			)

			deps.Logger.Warn("Post flagged as UNSAFE",
				zap.String("reason", moderation.Reason),
				zap.String("content_type", moderation.ContentType),
			)
		}

		// Finalize all steps: log, update, notify
		if err := finalizeModerationResult(deps, &payload, t, startedAt, status, message); err != nil {
			return err
		}

		return nil
	}
}

func finalizeModerationResult(
	deps FilterPostWorker,
	payload *FilterPostContentByAIPayload,
	task *asynq.Task,
	startedAt time.Time,
	status string,
	message string,
) error {
	// Save task log
	if err := updateTaskLog(deps, task, payload, startedAt, status, message); err != nil {
		deps.Logger.Error("Failed to update task log", zap.Error(err))
		return err
	}

	// Update post status
	if err := UpdatePublishPostResult(deps, payload.Post.ID.String(), status, message); err != nil {
		deps.Logger.Error("Failed to update post status", zap.Error(err))
		return err
	}

	// Send notification
	if err := deps.NotifyUser(&payload.Post, &payload.User, status, message); err != nil {
		deps.Logger.Error("Failed to notify user", zap.Error(err))
		// continue even if noti fails
	}

	return nil
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

	// log payload.Post.HTMLContent
	if payload.Post.HTMLContent == nil || *payload.Post.HTMLContent == "" {
		payload.Post.HTMLContent = new(string)
		*payload.Post.HTMLContent = "<p>No content provided</p>"
	}

	return fmt.Sprintf(`You are a strict AI content moderator tasked with reviewing human-written posts.

Your goal is to determine whether the content is SAFE or UNSAFE for publication.
You must respond in one of the following formats ONLY:
- SAFE: [REASON] | [CONTENT_TYPE]
- UNSAFE: [REASON with clear justification and quoted words/phrases] | [CONTENT_TYPE]

CONTENT_TYPE must be one of:
ARTICLE, NEWS, TUTORIAL, REVIEW, ADVERTISEMENT, ENTERTAINMENT, ADULT, SPAM, SEXUAL, VIOLENT, HATE, PROVOCATIVE, MISINFORMATION, POLITICAL, HEALTH, FINANCIAL, EDUCATIONAL, OTHER

### Evaluation Criteria:

1. **Profanity or Offensive Language**
   - Mark as UNSAFE if the content contains profanity, hate speech, Thai slang (e.g., 'มึง', 'เหี้ย', 'ควย') or English slurs (e.g., 'fuck', 'shit', 'wtf').
   - Always quote the detected words in your reason.

2. **Meaningless or Low-Value Content**
   - If the content is structured (with Introduction, Body, Conclusion, etc.) but lacks insight, explanation, or real information, mark as:
     > UNSAFE: Structured but lacks substance | OTHER

3. **Template or Repetitive Text**
   - If the content looks like filler or AI-generated boilerplate (e.g., "Technology is important for the future" without elaboration), mark as:
     > UNSAFE: Generic or repetitive phrases without value | OTHER

4. **Code Snippets Without Explanation**
   - If the post contains only code (HTML, Go, React, etc.) with little or no meaningful explanation or context, mark as:
     > UNSAFE: Code-only content without explanation | OTHER

5. **Valuable Content**
   - SAFE content should teach, inform, or explain in a clear and structured way with examples, facts, or opinions.

### DOs and DON'Ts

- ✅ DO focus on meaning, usefulness, and clarity.
- ❌ DO NOT mark as SAFE just because the content has formatting or code.
- ❌ DO NOT assume content is valuable unless it actually teaches or explains.
- ❌ DO NOT rely on structure alone (e.g., presence of headings or lists).
- ✅ DO look for explanations, reasoning, or insights that would benefit a human reader.

### Examples:

- SAFE: A clear step-by-step tutorial on Docker setup with explanation | TUTORIAL  
- UNSAFE: Repeats "Tech is the future" without explanation | OTHER  
- UNSAFE: Structured post but lacks actual informative content | OTHER  
- UNSAFE: Detected word 'เหี้ย' | HATE  
- SAFE: Detailed review with pros and cons of a product | REVIEW  
- UNSAFE: Just raw HTML/React/Go code without context | OTHER  
- UNSAFE: Contains phrase 'hot girls' or 'sexy content' | ADULT  
- UNSAFE: Post says only "ลองโพสต์เฉยๆครับ" or "test content" | OTHER

--- Content to Review ---

Title: %s  
Description: %s  

Content (in HTML format):  
%s

--- End ---`, payload.Post.Title, payload.Post.Description, *payload.Post.HTMLContent)
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

func (deps FilterPostWorker) NotifyUser(post *models.Post, user *models.User, status string, message string) error {
	// Prepare notification message
	formattedMessage := fmt.Sprintf("Your post moderation is complete.\n\nPost Title: %s\n\nStatus: %s\nResult: %s",
		post.Title,
		status,
		message,
	)

	// Determine notification title based on status
	notiTitle := "Post Moderation Result"
	if status == "SUCCESS" {
		notiTitle = "Your post has been published successfully"
	} else if status == "UNSAFE" {
		notiTitle = "Your post was rejected due to content policy violation"
	}

	err := deps.NotiService.Notify(
		user,
		notiTitle,
		"notification",
		formattedMessage,
		nil,
	)
	if err != nil {
		deps.Logger.Error("Failed to notify user about post moderation result",
			zap.String("user_id", user.ID.String()),
			zap.Error(err))
		return err
	}
	return nil
}
