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
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/post"
	"rag-searchbot-backend/pkg/ginctx"
	"rag-searchbot-backend/pkg/response"
	"rag-searchbot-backend/pkg/utils"
	"sort"
	"strconv"
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

const (
	DefaultTopK                = 10
	DefaultSimilarityThreshold = 0.35
	MaxTopK                    = 20
	MinSimilarityThreshold     = 0.1
)

type RAGConfig struct {
	TopK                int     `json:"top_k"`
	SimilarityThreshold float64 `json:"similarity_threshold"`
	Model               string  `json:"model"`
	Host                string  `json:"host"`
	UseSelfHost         bool    `json:"use_self_host"`
	APIKey              string  `json:"api_key"`
}

type ScoredChunk struct {
	Text  string  `json:"text"`
	Score float64 `json:"score"`
}

type StreamResponse struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

func (a *AIHandler) Chat(c *gin.Context) {
	// 1. Validate request and extract parameters
	req, postID, user, err := a.validateChatRequest(c)
	if err != nil {
		return // Error already handled in validateChatRequest
	}

	a.logger.Info("Chat handler triggered",
		zap.String("post_id", postID),
		zap.String("user_email", user.Email))

	// 2. Validate post and AI chat availability
	_, err = a.validatePost(c, postID)
	if err != nil {
		return // Error already handled in validatePost
	}

	// 3. Setup streaming response
	a.setupStreamingHeaders(c)

	// 4. Get RAG configuration
	config := a.getRAGConfig()

	// 5. Process RAG pipeline
	context, err := a.processRAGPipeline(c, postID, req.Prompt, config)
	if err != nil {
		return // Error already handled in processRAGPipeline
	}

	// 6. Generate and stream response
	a.generateAndStreamResponse(c, req.Prompt, context, config)
}

func (a *AIHandler) validateChatRequest(c *gin.Context) (*ChatRequestDTO, string, *models.User, error) {
	var req ChatRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		a.logger.Error("Invalid JSON format", zap.Error(err))
		a.writeErrorEvent(c, "Invalid JSON format")
		return nil, "", nil, err
	}

	postID := c.Param("post_id")
	if postID == "" {
		a.logger.Error("Missing post_id in request")
		a.writeErrorEvent(c, "Post ID required")
		return nil, "", nil, fmt.Errorf("missing post_id")
	}

	user, ok := ginctx.GetUserFromContext(c)
	if !ok || user == nil {
		a.logger.Error("Missing user context")
		response.JSONError(c, http.StatusUnauthorized, "Unauthorized", "User context missing")
		return nil, "", nil, fmt.Errorf("missing user context")
	}

	return &req, postID, user, nil
}

func (a *AIHandler) validatePost(c *gin.Context, postID string) (*models.Post, error) {
	post, err := a.PosRepo.GetByID(postID)
	if err != nil {
		a.logger.Error("Error fetching post", zap.Error(err))
		a.writeErrorEvent(c, "Post fetch error")
		return nil, err
	}

	if post == nil || !post.AIChatOpen {
		a.logger.Warn("Post not found or AI chat not enabled", zap.String("post_id", postID))
		a.writeErrorEvent(c, "Post not found or AI chat not enabled")
		return nil, fmt.Errorf("post not available")
	}

	return post, nil
}

func (a *AIHandler) setupStreamingHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Flush()
}

func (a *AIHandler) getRAGConfig() RAGConfig {
	config := RAGConfig{
		TopK:                DefaultTopK,
		SimilarityThreshold: DefaultSimilarityThreshold,
		Model:               os.Getenv("AI_MODEL"),
		Host:                os.Getenv("AI_HOST"),
		UseSelfHost:         os.Getenv("AI_SELF_HOST") == "true",
		APIKey:              os.Getenv("AI_API_KEY"),
	}

	// Parse custom TopK if provided
	if topKStr := os.Getenv("RAG_TOP_K"); topKStr != "" {
		if topK, err := strconv.Atoi(topKStr); err == nil && topK > 0 && topK <= MaxTopK {
			config.TopK = topK
		}
	}

	// Parse custom similarity threshold if provided
	if thresholdStr := os.Getenv("RAG_SIMILARITY_THRESHOLD"); thresholdStr != "" {
		if threshold, err := strconv.ParseFloat(thresholdStr, 64); err == nil &&
			threshold >= MinSimilarityThreshold && threshold <= 1.0 {
			config.SimilarityThreshold = threshold
		}
	}

	return config
}

func (a *AIHandler) processRAGPipeline(c *gin.Context, postID, question string, config RAGConfig) (string, error) {
	a.logger.Info("Processing RAG pipeline", zap.String("question", question))

	// 1. Get question embedding
	questionEmbedding32, err := ai.GetEmbedding(question)
	if err != nil {
		a.logger.Error("Failed to get question embedding", zap.Error(err))
		a.writeErrorEvent(c, "Embedding error")
		return "", err
	}

	// Convert []float32 to []float64
	questionEmbedding := make([]float64, len(questionEmbedding32))
	for i, v := range questionEmbedding32 {
		questionEmbedding[i] = float64(v)
	}

	// 2. Retrieve and score chunks
	scoredChunks, err := a.retrieveAndScoreChunks(postID, questionEmbedding)
	if err != nil {
		a.logger.Error("Failed to retrieve chunks", zap.Error(err))
		a.writeErrorEvent(c, "Failed to retrieve relevant content")
		return "", err
	}

	// 3. Select top chunks
	selectedChunks := a.selectTopChunks(scoredChunks, config)

	// 4. Build context
	context := a.buildContext(selectedChunks)

	a.logger.Info("RAG pipeline completed",
		zap.Int("total_chunks", len(scoredChunks)),
		zap.Int("selected_chunks", len(selectedChunks)),
		zap.Int("context_length", len(context)))

	return context, nil
}

func (a *AIHandler) retrieveAndScoreChunks(postID string, questionEmbedding []float64) ([]ScoredChunk, error) {
	embeddedChunks, err := a.PosRepo.GetEmbeddingByPostID(postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post embeddings: %w", err)
	}

	if len(embeddedChunks) == 0 {
		return nil, fmt.Errorf("no embedded chunks found for post")
	}

	scoredChunks := make([]ScoredChunk, 0, len(embeddedChunks))
	for _, chunk := range embeddedChunks {
		vec := chunk.Vector.Slice()

		questionEmbedding32 := make([]float32, len(questionEmbedding))
		for i, v := range questionEmbedding {
			questionEmbedding32[i] = float32(v)
		}

		score := utils.CosineSimilarity(vec, questionEmbedding32)
		scoredChunks = append(scoredChunks, ScoredChunk{
			Text:  chunk.Content,
			Score: score,
		})

		a.logger.Debug("Chunk scored",
			zap.Float64("score", score),
			zap.String("preview", a.truncateText(chunk.Content, 50)))
	}

	// Sort by score descending
	sort.Slice(scoredChunks, func(i, j int) bool {
		return scoredChunks[i].Score > scoredChunks[j].Score
	})

	return scoredChunks, nil
}

func (a *AIHandler) selectTopChunks(scoredChunks []ScoredChunk, config RAGConfig) []ScoredChunk {
	if len(scoredChunks) == 0 {
		return nil
	}

	var selectedChunks []ScoredChunk

	// Strategy 1: Take top K chunks regardless of score
	topK := config.TopK
	if topK > len(scoredChunks) {
		topK = len(scoredChunks)
	}

	for i := 0; i < topK; i++ {
		selectedChunks = append(selectedChunks, scoredChunks[i])
	}

	// Strategy 2: Add additional chunks above similarity threshold
	// (but avoid duplicates from the top K selection)
	for i := topK; i < len(scoredChunks); i++ {
		if scoredChunks[i].Score >= config.SimilarityThreshold {
			selectedChunks = append(selectedChunks, scoredChunks[i])
		} else {
			break // Since it's sorted, no need to check further
		}
	}

	// Log selection statistics
	if len(selectedChunks) > 0 {
		a.logger.Info("Chunk selection completed",
			zap.Int("selected_count", len(selectedChunks)),
			zap.Float64("highest_score", selectedChunks[0].Score),
			zap.Float64("lowest_score", selectedChunks[len(selectedChunks)-1].Score),
			zap.Float64("threshold", config.SimilarityThreshold))
	}

	return selectedChunks
}

func (a *AIHandler) buildContext(chunks []ScoredChunk) string {
	if len(chunks) == 0 {
		return ""
	}

	contextParts := make([]string, len(chunks))
	for i, chunk := range chunks {
		// Optionally include score in context (useful for debugging)
		// contextParts[i] = fmt.Sprintf("[Score: %.3f] %s", chunk.Score, chunk.Text)
		contextParts[i] = chunk.Text
	}

	return strings.Join(contextParts, "\n\n")
}

func (a *AIHandler) buildSystemPrompt(context string) string {
	if strings.TrimSpace(context) == "" {
		a.logger.Warn("No context provided, using fallback system message")
		return `❗ ไม่พบข้อมูลในบทความนี้ / No relevant content found in the article.`
	}

	return fmt.Sprintf(`คุณเป็นผู้ช่วย AI ที่ตอบคำถามโดยใช้ข้อมูลจากบทความเท่านั้น:
(You are an AI assistant that must answer using **only** the content provided below.)

-----
%s
-----

ห้ามตอบจากความรู้ของคุณเองโดยเด็ดขาด หากไม่มีข้อมูลที่เกี่ยวข้องในบทความ กรุณาตอบว่า "ไม่พบข้อมูลในบทความนี้"
(Do **not** use your own general knowledge. If the answer cannot be found in the content above, reply: "No information available in the article.")`, context)
}

func (a *AIHandler) generateAndStreamResponse(c *gin.Context, question, context string, config RAGConfig) {
	systemPrompt := a.buildSystemPrompt(context)

	payload := map[string]interface{}{
		"model":  config.Model,
		"stream": true,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": question},
		},
	}

	a.logger.Debug("Sending LLM request",
		zap.String("model", config.Model),
		zap.String("question", question),
		zap.String("context_preview", a.truncateText(context, 100)))

	resp, err := a.sendLLMRequest(payload, config)
	if err != nil {
		a.logger.Error("LLM service error", zap.Error(err))
		a.writeErrorEvent(c, "LLM service error")
		return
	}
	defer resp.Body.Close()

	a.writeEvent(c, "start", "Streaming started")
	a.streamLLMResponse(c, resp)
}

func (a *AIHandler) sendLLMRequest(payload map[string]interface{}, config RAGConfig) (*http.Response, error) {
	body, _ := json.Marshal(payload)

	if config.UseSelfHost {
		a.logger.Info("Using self-hosted Ollama", zap.String("host", config.Host))
		return http.Post(config.Host+"/api/chat", "application/json", bytes.NewBuffer(body))
	}

	// Using OpenRouter
	a.logger.Info("Using OpenRouter")
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenRouter API key missing")
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	req.Header.Set("HTTP-Referer", "https://blog.bsospace.com")
	req.Header.Set("X-Title", "https://blog.bsospace.com")

	client := &http.Client{}
	return client.Do(req)
}

func (a *AIHandler) streamLLMResponse(c *gin.Context, resp *http.Response) {
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				a.logger.Info("LLM stream finished")
				a.writeEvent(c, "end", "done")
			} else {
				a.logger.Error("Error reading LLM stream", zap.Error(err))
				a.writeErrorEvent(c, "Stream reading error")
			}
			break
		}

		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		raw := bytes.TrimSpace(line[6:])
		if len(raw) == 0 {
			continue
		}

		if bytes.Equal(raw, []byte("[DONE]")) {
			a.writeEvent(c, "end", "done")
			break
		}

		if content := a.parseStreamChunk(raw); content != "" {
			jsonEncoded, _ := json.Marshal(map[string]string{"text": content})
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonEncoded)
			c.Writer.Flush()
		}
	}
}

func (a *AIHandler) parseStreamChunk(raw []byte) string {
	var chunk map[string]interface{}
	if err := json.Unmarshal(raw, &chunk); err != nil {
		a.logger.Warn("Failed to parse chunk", zap.String("raw", string(raw)), zap.Error(err))
		return ""
	}

	// Handle Ollama format
	if message, ok := chunk["message"].(map[string]interface{}); ok {
		if content, ok := message["content"].(string); ok {
			return content
		}
	}

	// Handle OpenAI/OpenRouter format
	if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
		choice := choices[0].(map[string]interface{})
		if delta, ok := choice["delta"].(map[string]interface{}); ok {
			if content, ok := delta["content"].(string); ok {
				return content
			}
		}
	}

	return ""
}

// Helper methods
func (a *AIHandler) writeEvent(c *gin.Context, event, data string) {
	fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data)
	c.Writer.Flush()
}

func (a *AIHandler) writeErrorEvent(c *gin.Context, message string) {
	fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", message)
	c.Writer.Flush()
}

func (a *AIHandler) truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
