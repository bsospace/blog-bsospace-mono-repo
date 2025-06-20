package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"rag-searchbot-backend/internal/ai"
	"rag-searchbot-backend/internal/post"
	"rag-searchbot-backend/pkg/ginctx"
	"rag-searchbot-backend/pkg/response"
	"rag-searchbot-backend/pkg/utils"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AIHandler struct {
	AIService *ai.AIService
	PosRepo   post.PostRepositoryInterface
	logger    *zap.Logger
}

func NewAIHandler(aiService *ai.AIService, posRepo post.PostRepositoryInterface, logger *zap.Logger) *AIHandler {
	return &AIHandler{
		AIService: aiService,
		PosRepo:   posRepo,
		logger:    logger,
	}
}

func (a *AIHandler) OpenAIMode(c *gin.Context) {

	postID := c.Param("post_id")

	if postID == "" {
		response.JSONSuccess(c, http.StatusBadRequest, "Bad request", "Post id required")
		return
	}

	user, ok := ginctx.GetUserFromContext(c)
	if !ok || user == nil {
		response.JSONError(c, http.StatusUnauthorized, "User not found in context", "User context is missing")
	}

	// Use existingPost in your further logic here
	result, err := a.AIService.OpenAIMode(postID, user)

	if err != nil && err == gorm.ErrRecordNotFound {
		response.JSONError(c, http.StatusNotFound, "Not found", "Post not found!")
		return
	}

	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	if result {
		response.JSONSuccess(c, http.StatusOK, "Success", "AI mode in queue")
		return
	}

	response.JSONSuccess(c, http.StatusOK, "Success", "AI mode disabled")
}

func (a *AIHandler) DisableOpenAIMode(c *gin.Context) {
	postID := c.Param("post_id")

	if postID == "" {
		response.JSONSuccess(c, http.StatusBadRequest, "Bad request", "Post id required")
		return
	}

	user, ok := ginctx.GetUserFromContext(c)
	if !ok || user == nil {
		response.JSONError(c, http.StatusUnauthorized, "User not found in context", "User context is missing")
	}

	result, err := a.AIService.DisableOpenAIMode(postID, user)

	if err != nil && err == gorm.ErrRecordNotFound {
		response.JSONError(c, http.StatusNotFound, "Not found", "Post not found!")
		return
	}

	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	if result {
		response.JSONSuccess(c, http.StatusOK, "Success", "AI mode disabled")
		return
	}

	response.JSONSuccess(c, http.StatusOK, "Success", "AI mode already disabled")
}

type ChatRequestDTO struct {
	Prompt string `json:"prompt"`
}

type AskRequest struct {
	Question string `json:"question"`
}

// Refactored Chat method using pre-computed vector embeddings from the database
// Refactored Chat method using dynamic embedding from post content
func (a *AIHandler) Chat(c *gin.Context) {
	var req ChatRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("Invalid JSON format", zap.String("body", c.Request.URL.Path), zap.Error(err))
		response.JSONError(c, http.StatusBadRequest, "Bad request", "Invalid JSON format")
		return
	}

	postID := c.Param("post_id")
	if postID == "" {
		a.logger.Error("Missing post_id in request")
		response.JSONError(c, http.StatusBadRequest, "Bad request", "Post id required")
		return
	}

	user, ok := ginctx.GetUserFromContext(c)
	if !ok || user == nil {
		a.logger.Error("Missing user context")
		response.JSONError(c, http.StatusUnauthorized, "Unauthorized", "User context missing")
		return
	}

	a.logger.Info("Chat handler triggered", zap.String("post_id", postID), zap.String("user_email", user.Email))

	post, err := a.PosRepo.GetByID(postID)
	if err != nil {
		a.logger.Error("Error fetching post", zap.Error(err))
		response.JSONError(c, http.StatusInternalServerError, "Internal error", "Post fetch error")
		return
	}
	if post == nil || !post.AIChatOpen {
		a.logger.Warn("Post not found or AI chat not enabled", zap.String("post_id", postID))
		response.JSONError(c, http.StatusNotFound, "Post not found or AI chat not enabled", "")
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	prompt := AskRequest{Question: req.Prompt}
	a.logger.Info("Prompt parsed", zap.String("question", prompt.Question))

	questionEmbedding, err := ai.GetEmbedding(prompt.Question)
	if err != nil {
		a.logger.Error("Failed to get embedding", zap.Error(err))
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", "Embedding error")
		c.Writer.Flush()
		return
	}

	type ScoredChunk struct {
		Text  string
		Score float64
	}

	// ดึง embeddings ที่สร้างไว้แล้ว (จาก DB)
	embeddedChunks, err := a.PosRepo.GetEmbeddingByPostID(postID)
	if err != nil {
		a.logger.Error("Failed to get post embeddings", zap.Error(err))
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", "Failed to get post embeddings")
		c.Writer.Flush()
		return
	}

	var scoredChunks []ScoredChunk
	for _, chunk := range embeddedChunks {
		score := utils.CosineSimilarity(chunk.Vector.Slice(), questionEmbedding)
		scoredChunks = append(scoredChunks, ScoredChunk{Text: chunk.Content, Score: score})
		a.logger.Debug("Chunk score", zap.Float64("score", score), zap.String("text", chunk.Content))
	}

	if len(scoredChunks) == 0 {
		a.logger.Warn("No chunks found for the post", zap.String("post_id", postID))
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", "No relevant information found in the post")
		c.Writer.Flush()
		return
	}

	sort.Slice(scoredChunks, func(i, j int) bool {
		return scoredChunks[i].Score > scoredChunks[j].Score
	})

	topChunks := []string{}
	for i := 0; i < len(scoredChunks) && len(topChunks) < 10; i++ {
		topChunks = append(topChunks, scoredChunks[i].Text)
	}

	for _, chunk := range scoredChunks {
		if chunk.Score > 0.35 {
			topChunks = append(topChunks, chunk.Text)
		}
	}

	fullContext := strings.Join(topChunks, "\n\n")
	a.logger.Info("Full context for LLM", zap.String("context", fullContext), zap.Int("chunk_count", len(topChunks)))

	systemContext := ""
	if strings.TrimSpace(fullContext) == "" {
		a.logger.Warn("No context provided, using strict fallback system message")
		systemContext = `❗ ไม่พบข้อมูลในบทความนี้ / No relevant content found in the article.`
	} else {
		systemContext = `คุณเป็นผู้ช่วย AI ที่ตอบคำถามโดยใช้ข้อมูลจากบทความเท่านั้น:
(You are an AI assistant that must answer using **only** the content provided below.)

-----
` + fullContext + `
-----

ห้ามตอบจากความรู้ของคุณเองโดยเด็ดขาด หากไม่มีข้อมูลที่เกี่ยวข้องในบทความ กรุณาตอบว่า "ไม่พบข้อมูลในบทความนี้"
(Do **not** use your own general knowledge. If the answer cannot be found in the content above, reply: "No information available in the article.")`
	}

	model := os.Getenv("AI_MODEL")
	host := os.Getenv("AI_HOST")
	useSelfHost := os.Getenv("AI_SELF_HOST") == "true"

	payload := map[string]interface{}{
		"model":  model,
		"stream": true,
		"messages": []map[string]string{
			{"role": "system", "content": systemContext},
			{"role": "user", "content": prompt.Question},
		},
	}
	body, _ := json.Marshal(payload)

	var resp *http.Response
	if useSelfHost {
		a.logger.Info("Using self-hosted Ollama")
		resp, err = http.Post(host+"/api/chat", "application/json", bytes.NewBuffer(body))
	} else {
		a.logger.Info("Using OpenRouter")
		openRouterAPIKey := os.Getenv("AI_API_KEY")
		if openRouterAPIKey == "" {
			a.logger.Error("Missing OpenRouter API Key")
			fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", "OpenRouter API Key missing")
			c.Writer.Flush()
			return
		}
		reqOpenRouter, _ := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(body))
		reqOpenRouter.Header.Set("Content-Type", "application/json")
		reqOpenRouter.Header.Set("Authorization", "Bearer "+openRouterAPIKey)
		client := &http.Client{}
		resp, err = client.Do(reqOpenRouter)

		a.logger.Info("OpenRouter request sent", zap.String("url", reqOpenRouter.URL.String()))
	}

	if err != nil {
		a.logger.Error("LLM service error", zap.Error(err))
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", "LLM service error")
		c.Writer.Flush()
		return
	}
	defer resp.Body.Close()

	fmt.Fprintf(c.Writer, "event: start\ndata: %s\n\n", "Streaming started")
	c.Writer.Flush()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				a.logger.Info("LLM stream finished (EOF)")
			} else {
				a.logger.Error("Error reading LLM stream", zap.Error(err))
			}
			break
		}
		if bytes.HasPrefix(line, []byte("data: ")) {
			raw := bytes.TrimSpace(line[6:])
			if len(raw) == 0 {
				continue
			}
			if bytes.Equal(raw, []byte("[DONE]")) {
				fmt.Fprintf(c.Writer, "event: end\ndata: %s\n\n", "done")
				c.Writer.Flush()
				break
			}
			var chunk map[string]interface{}
			if err := json.Unmarshal(raw, &chunk); err != nil {
				a.logger.Warn("Failed to parse chunk", zap.String("raw", string(raw)), zap.Error(err))
				continue
			}
			if message, ok := chunk["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					jsonEncoded, _ := json.Marshal(map[string]string{"text": content})
					fmt.Fprintf(c.Writer, "data: %s\n\n", jsonEncoded)
					c.Writer.Flush()
				}
			} else if delta, ok := chunk["choices"].([]interface{}); ok {
				if len(delta) > 0 {
					choice := delta[0].(map[string]interface{})
					if deltaContent, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := deltaContent["content"].(string); ok {
							jsonEncoded, _ := json.Marshal(map[string]string{"text": content})
							fmt.Fprintf(c.Writer, "data: %s\n\n", jsonEncoded)
							c.Writer.Flush()
						}
					}
				}
			}
		}
	}
}

func splitText(text string, chunkSize, overlap int) []string {
	var chunks []string
	runes := []rune(text)
	for i := 0; i < len(runes); i += chunkSize - overlap {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
		if end == len(runes) {
			break
		}
	}
	return chunks
}
