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

	"rag-searchbot-backend/pkg/token"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	StrictThreshold            = 0.4
)

type RAGConfig struct {
	TopK                int     `json:"top_k"`
	SimilarityThreshold float64 `json:"similarity_threshold"` // ใช้ใน logic เดิม
	StrictThreshold     float64 `json:"strict_threshold"`     // สำหรับการกรองแบบเข้มงวด
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
	a.generateAndStreamResponse(c, req.Prompt, context, config, postID, user)
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
	origin := c.Request.Header.Get("Origin")

	appEnv := os.Getenv("APP_ENV")
	var allowedOrigins []string
	if appEnv == "release" {
		allowedOrigins = strings.Split(os.Getenv("ALLOWED_ORIGINS_PROD"), ",")
	} else {
		allowedOrigins = strings.Split(os.Getenv("ALLOWED_ORIGINS_DEV"), ",")
	}

	// Trim และตรวจสอบว่า origin อยู่ใน allowed list หรือไม่
	isAllowed := false
	for _, o := range allowedOrigins {
		if strings.TrimSpace(o) == origin {
			isAllowed = true
			break
		}
	}

	if isAllowed {
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()
}

func (a *AIHandler) getRAGConfig() RAGConfig {
	config := RAGConfig{
		TopK:                DefaultTopK,
		SimilarityThreshold: DefaultSimilarityThreshold,
		StrictThreshold:     StrictThreshold,
		Model:               os.Getenv("AI_MODEL"),
		Host:                os.Getenv("AI_HOST"),
		UseSelfHost:         os.Getenv("AI_SELF_HOST") == "true",
		APIKey:              os.Getenv("AI_API_KEY"),
	}

	// Parse RAG_TOP_K
	if topKStr := os.Getenv("RAG_TOP_K"); topKStr != "" {
		if topK, err := strconv.Atoi(topKStr); err == nil && topK > 0 && topK <= MaxTopK {
			config.TopK = topK
		}
	}

	// Parse RAG_SIMILARITY_THRESHOLD
	if thresholdStr := os.Getenv("RAG_SIMILARITY_THRESHOLD"); thresholdStr != "" {
		if threshold, err := strconv.ParseFloat(thresholdStr, 64); err == nil {
			config.SimilarityThreshold = threshold
		}
	}

	// Parse RAG_STRICT_THRESHOLD
	if strictStr := os.Getenv("RAG_STRICT_THRESHOLD"); strictStr != "" {
		if strict, err := strconv.ParseFloat(strictStr, 64); err == nil {
			config.StrictThreshold = strict
		}
	}

	return config
}

func (a *AIHandler) processRAGPipeline(c *gin.Context, postID, question string, config RAGConfig) (string, error) {
	a.logger.Info("Processing RAG pipeline", zap.String("question", question))

	allScoredChunks := []ScoredChunk{}
	phrases := SplitQuestionToPhrases(question)

	for _, phrase := range phrases {
		embedding32, err := ai.GetEmbedding(phrase)
		if err != nil {
			a.logger.Warn("Failed to get embedding for phrase", zap.String("phrase", phrase), zap.Error(err))
			continue
		}

		// log the embedding for debugging
		a.logger.Debug("Embedding generated",
			zap.String("phrase", phrase))

		embedding := make([]float64, len(embedding32))
		for i, v := range embedding32 {
			embedding[i] = float64(v)
		}

		scoredChunks, err := a.retrieveAndScoreChunks(postID, embedding)
		if err != nil {
			a.logger.Warn("Chunk scoring failed for phrase", zap.String("phrase", phrase), zap.Error(err))
			continue
		}

		// log the number of chunks scored
		a.logger.Debug("Scored chunks retrieved",
			zap.String("phrase", phrase),
			zap.Int("chunk_count", len(scoredChunks)))

		allScoredChunks = append(allScoredChunks, scoredChunks...)
	}

	if len(allScoredChunks) == 0 {
		a.logger.Warn("No relevant chunks found after processing all phrases")
		a.writeErrorEvent(c, "No relevant content found")
		return "", nil
	}

	// Remove duplicates by content text (optional)
	uniqueMap := map[string]ScoredChunk{}
	for _, chunk := range allScoredChunks {
		if _, ok := uniqueMap[chunk.Text]; !ok {
			uniqueMap[chunk.Text] = chunk
		}
	}

	// Convert back to slice and sort
	uniqueChunks := make([]ScoredChunk, 0, len(uniqueMap))
	for _, chunk := range uniqueMap {
		uniqueChunks = append(uniqueChunks, chunk)
	}
	sort.Slice(uniqueChunks, func(i, j int) bool {
		return uniqueChunks[i].Score > uniqueChunks[j].Score
	})

	// Select top relevant ones
	selectedChunks := a.selectTopChunks(uniqueChunks, config)

	// Debug
	// for i, chunk := range selectedChunks {
	// 	a.logger.Debug("Selected chunk",
	// 		zap.Int("index", i),
	// 		zap.Float64("score", chunk.Score),
	// 		zap.String("text", chunk.Text))
	// }

	context := a.buildContext(selectedChunks)

	a.logger.Info("RAG pipeline completed",
		zap.Int("total_chunks", len(uniqueChunks)),
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
	for _, chunk := range scoredChunks {
		if chunk.Score >= config.StrictThreshold {
			selectedChunks = append(selectedChunks, chunk)
		}
	}

	// log the number of chunks selected by strict filter
	a.logger.Debug("Chunks selected by strict filter",
		zap.Int("count", len(selectedChunks)),
		zap.Float64("strict_threshold", config.StrictThreshold))

	// ถ้าไม่มีเลย fallback ไปใช้ top K
	if len(selectedChunks) == 0 {
		topK := config.TopK
		if topK > len(scoredChunks) {
			topK = len(scoredChunks)
		}
		selectedChunks = scoredChunks[:topK]

		a.logger.Warn("Strict filter empty, fallback to top K",
			zap.Int("top_k", topK))
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
		return `หากเป็นคำทักทาย เช่น "สวัสดี", "สวัสดีค่ะ", "ไงบ้าง", "เฮลโหล"
ให้ตอบกลับอย่างสุภาพและเป็นมิตร เช่น "สวัสดีค่ะ ยินดีให้บริการค่ะ 😊"

หากไม่ใช่คำทักทาย และไม่มีข้อมูลบทความ ให้ตอบว่า
"ไม่พบข้อมูลในบทความนี้ค่ะ ลองถามคำถามอื่นได้นะคะ 😊"
`
	}

	return fmt.Sprintf(`คุณเป็นผู้ช่วยสรุปเนื้อหาจากบทความเท่านั้น ห้ามใช้ความรู้ภายนอกเด็ดขาด

หากผู้ใช้พิมพ์คำทักทาย เช่น "สวัสดี", "เฮลโหล", "ไง", "หวัดดี", "hi", "hello"  
ให้ตอบกลับอย่างสุภาพ เช่น "สวัสดีค่ะ ยินดีให้บริการนะคะ 😊"

หากผู้ใช้ถามคำถาม ให้ตอบโดยอิงจากเนื้อหาด้านล่างนี้เท่านั้น  
หากไม่มีข้อมูลเกี่ยวข้อง ให้ตอบว่า  
"ไม่พบข้อมูลในบทความนี้ค่ะ ลองถามคำถามอื่นได้นะคะ 😊"

-----  
%s  
-----`, context)
}

func (a *AIHandler) generateAndStreamResponse(c *gin.Context, question, context string, config RAGConfig, postID string, user *models.User) {
	systemPrompt := a.buildSystemPrompt(context)
	inputText := systemPrompt + "\n" + question
	inputTokens := token.CountTokens(inputText)

	maxContextTokens, _ := strconv.Atoi(os.Getenv("AI_MAX_TOKENS"))
	if maxContextTokens == 0 {
		maxContextTokens = 3000
	}

	maxNewTokens := maxContextTokens - inputTokens
	if maxNewTokens < 256 {
		maxNewTokens = 256
	} else if maxNewTokens > 1024 {
		maxNewTokens = 1024
	}

	payload := map[string]interface{}{
		"model":      config.Model,
		"stream":     true,
		"max_tokens": maxNewTokens,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": question},
		},
	}

	resp, err := a.sendLLMRequest(payload, config)
	if err != nil || resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		a.logger.Error("LLM error", zap.Error(err), zap.String("body", string(bodyBytes)))
		a.writeErrorEvent(c, "LLM service error")
		return
	}
	defer resp.Body.Close()

	postUUID, err := uuid.Parse(postID)
	if err != nil {
		a.writeErrorEvent(c, "Invalid post ID format")
		return
	}

	a.writeEvent(c, "start", "Streaming started")

	// Stream the response
	fullText := a.streamLLMResponse(c, resp)

	// Get real token usage
	var realTotalTokens int
	if !config.UseSelfHost {
		// clone payload และเปลี่ยน stream = false
		usagePayload := map[string]interface{}{
			"model":      config.Model,
			"stream":     false,
			"max_tokens": maxNewTokens,
			"messages": []map[string]string{
				{"role": "system", "content": systemPrompt},
				{"role": "user", "content": question},
			},
		}

		usageResp, err := a.sendLLMRequest(usagePayload, config)
		if err == nil && usageResp.StatusCode == http.StatusOK {
			defer usageResp.Body.Close()
			var result map[string]interface{}
			if err := json.NewDecoder(usageResp.Body).Decode(&result); err == nil {
				if usage, ok := result["usage"].(map[string]interface{}); ok {
					promptTokens := int(usage["prompt_tokens"].(float64))
					completionTokens := int(usage["completion_tokens"].(float64))
					realTotalTokens = promptTokens + completionTokens

					a.logger.Info("Token usage from OpenRouter",
						zap.Int("prompt_tokens", promptTokens),
						zap.Int("completion_tokens", completionTokens),
						zap.Int("total_tokens", realTotalTokens))
				}
			}
		}
	}

	// Fallback to input tokens + full text tokens if real total tokens not available
	if realTotalTokens == 0 {
		realTotalTokens = inputTokens + token.CountTokens(fullText)
	}

	// Log the final token usage
	if err := a.SaveChatHistory(c, &models.Post{ID: postUUID}, user, fullText, question, realTotalTokens); err != nil {
		a.logger.Error("Failed to save chat history", zap.Error(err))
		a.writeErrorEvent(c, "Failed to save chat history")
	}
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

func (a *AIHandler) streamLLMResponse(c *gin.Context, resp *http.Response) string {
	reader := bufio.NewReader(resp.Body)
	var fullText strings.Builder

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

		content := a.parseStreamChunk(raw)
		if content != "" {
			fullText.WriteString(content)
			jsonEncoded, _ := json.Marshal(map[string]string{"text": content})
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonEncoded)
			c.Writer.Flush()
		}
	}

	return fullText.String()
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

func SplitText(text string, chunkSize int, overlap int) []string {
	var chunks []string
	for start := 0; start < len(text); start += chunkSize - overlap {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[start:end])
		if end == len(text) {
			break
		}
	}
	return chunks
}

func SplitQuestionToPhrases(q string) []string {
	words := strings.Fields(q)

	// กรณีคำเดียวหรือไม่มีคำ
	if len(words) <= 1 {
		return words
	}

	// bi-gram phrase
	var phrases []string
	for i := 0; i < len(words)-1; i++ {
		phrases = append(phrases, strings.Join(words[i:i+2], " "))
	}
	return phrases
}

// save chat history
func (a *AIHandler) SaveChatHistory(c *gin.Context, post *models.Post, user *models.User, responseText string, promt string, tokenUse int) error {
	chat := &models.AIResponse{
		Response:  responseText,
		Prompt:    promt,
		UserID:    user.ID,
		PostID:    post.ID,
		TokenUsed: tokenUse,
		Success:   true,
		Model:     os.Getenv("AI_MODEL"),
	}

	if err := a.AIService.CreateChat(chat, post.ID.String(), user); err != nil {
		a.logger.Error("Failed to save chat history", zap.Error(err))
		return fmt.Errorf("failed to save chat history: %w", err)
	}

	a.logger.Info("Chat history saved successfully",
		zap.String("post_id", post.ID.String()),
		zap.String("response_text", responseText),
		zap.String("user_email", user.Email))

	return nil
}
