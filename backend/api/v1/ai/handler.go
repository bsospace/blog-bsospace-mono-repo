package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"rag-searchbot-backend/internal/ai"
	"rag-searchbot-backend/internal/llm"
	"rag-searchbot-backend/internal/llm_types"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/post"
	"rag-searchbot-backend/pkg/ginctx"
	"rag-searchbot-backend/pkg/response"
	"rag-searchbot-backend/pkg/tiptap"
	"rag-searchbot-backend/pkg/token"
	"rag-searchbot-backend/pkg/utils"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AIHandler struct {
	AIService                      *ai.AIService
	AgentIntentClassifierService   ai.AgentIntentClassifierServiceInterface
	PosRepo                        post.PostRepositoryInterface
	logger                         *zap.Logger
	agentAgentToolWebSearchService ai.AgentToolWebSearch
	llmClient                      llm.LLM
}

func NewAIHandler(aiService *ai.AIService,
	agentIntentClassifierService ai.AgentIntentClassifierServiceInterface,
	posRepo post.PostRepositoryInterface, logger *zap.Logger,
	agentAgentToolWebSearchService ai.AgentToolWebSearch, llmClient llm.LLM) *AIHandler {
	return &AIHandler{
		AIService:                      aiService,
		PosRepo:                        posRepo,
		logger:                         logger,
		AgentIntentClassifierService:   agentIntentClassifierService,
		agentAgentToolWebSearchService: agentAgentToolWebSearchService,
		llmClient:                      llmClient,
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
	DefaultTopK                = 10   // จำนวน context สูงสุดที่ดึงกลับมาต่อครั้ง (Top-K retrieval)
	DefaultSimilarityThreshold = 0.35 // ใช้เป็น threshold ปกติ ถ้าไม่มี override
	MaxTopK                    = 20   // กันไม่ให้ดึงเกินนี้แม้จะ fuzzy
	MinSimilarityThreshold     = 0.1  // ล่างสุดที่ยอมให้ผ่าน (fuzzy จริง ๆ)
	StrictThreshold            = 0.7  // ใช้ตอนต้องการ “เนื้อหาที่มั่นใจมาก” เท่านั้น
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

	a.logger.Info("===================================================== Chat handler triggered =====================================================",
		zap.String("post_id", postID),
		zap.String("user_email", user.Email))

	// Classify the content
	if strings.TrimSpace(req.Prompt) == "" {
		a.logger.Warn("Empty prompt received")
		response.JSONError(c, http.StatusBadRequest, "Bad request", "Prompt cannot be empty")
		return
	}

	post, err := a.validatePost(c, postID)

	a.logger.Info("Post validated for AI chat",
		zap.String("post_id", postID),
		zap.String("post_title", post.Title))
	if err != nil {
		return // Error already handled in validatePost
	}

	a.logger.Info("Post validated for AI chat", zap.String("post_id", postID),
		zap.String("post_title", post.Title))

	// classifyContent
	intent, err := a.AgentIntentClassifierService.ClassifyIntent(req.Prompt, post)
	if err != nil {
		a.logger.Warn("Failed to classify content", zap.Error(err))
		response.JSONError(c, http.StatusInternalServerError, "Internal server error", "Failed to classify content")
		return
	}

	a.logger.Info("Content classified",
		zap.String("intent", string(intent)))

	// classifyContent เดิม
	// intent, err := a.AIService.ClassifyContent(req.Prompt)
	// if err != nil {
	// 	a.logger.Warn("Failed to classify content", zap.Error(err))
	// 	// Continue with the rest of the pipeline even if classification fails
	// }

	a.logger.Info("Content classified",
		zap.String("intent_raw", fmt.Sprintf("%q", string(intent))),
		zap.String("intent", string(intent)))

	// 3. Setup streaming response
	a.setupStreamingHeaders(c)

	plaintextContent := tiptap.ExtractTextFromTiptap(post.Content)

	if string(intent) == "summarize_post" {
		// log the intent
		a.logger.Info("Intent detected: summarize_post",
			zap.String("post_id", postID),
			zap.String("prompt", req.Prompt),
			zap.String("plaintextContent", plaintextContent))

		fullText, err := a.AIService.StreamPostSummary(c.Request.Context(), req.Prompt, plaintextContent, func(chunk string) {
			jsonEncoded, _ := json.Marshal(map[string]string{"text": chunk})
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonEncoded)
			c.Writer.Flush()
		})
		if err != nil {
			a.logger.Error("Failed to stream post summary", zap.Error(err))
			a.writeErrorEvent(c, "Failed to stream post summary")
			return
		}

		// Save history
		postUUID, err := uuid.Parse(postID)
		if err != nil {
			a.logger.Error("Failed to parse post ID", zap.Error(err))
			return
		}

		// Save user question and AI response to history
		tokenCount := token.CountTokens(req.Prompt + fullText)
		if err := a.SaveChatHistory(c, &models.Post{ID: postUUID}, user, fullText, req.Prompt, tokenCount, os.Getenv("AI_MODEL")); err != nil {
			a.logger.Error("Failed to save chat history", zap.Error(err))
			return
		}
		return
	}

	if string(intent) == "greeting_farewell" {
		// log the intent
		a.logger.Info("Intent detected: greeting_farewell",
			zap.String("post_id", postID),
			zap.String("prompt", req.Prompt))

		fullText, err := a.AIService.StreamGreetingFarewell(c.Request.Context(), req.Prompt, func(chunk string) {
			jsonEncoded, _ := json.Marshal(map[string]string{"text": chunk})
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonEncoded)
			c.Writer.Flush()
		})
		if err != nil {
			a.logger.Error("Failed to stream greeting/farewell", zap.Error(err))
			a.writeErrorEvent(c, "Failed to stream greeting/farewell")
			return
		}

		// Save history
		postUUID, err := uuid.Parse(postID)
		if err != nil {
			a.logger.Error("Failed to parse post ID", zap.Error(err))
			return
		}

		// Save user question and AI response to history
		tokenCount := token.CountTokens(req.Prompt + fullText)
		if err := a.SaveChatHistory(c, &models.Post{ID: postUUID}, user, fullText, req.Prompt, tokenCount, os.Getenv("AWS_BEDROCK_LLM_MODEL")); err != nil {
			a.logger.Error("Failed to save chat history", zap.Error(err))
			return
		}
	}

	if string(intent) == "blog_question" {

		// log the intent
		a.logger.Info("Intent detected: blog_question",
			zap.String("post_id", postID),
			zap.String("prompt", req.Prompt))
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
	if string(intent) == "unknown" {

		// log the intent
		a.logger.Info("Intent detected: greeting_farewell",
			zap.String("post_id", postID),
			zap.String("prompt", req.Prompt))

		fullText, err := a.AIService.StreamGreetingFarewell(c.Request.Context(), req.Prompt, func(chunk string) {
			jsonEncoded, _ := json.Marshal(map[string]string{"text": chunk})
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonEncoded)
			c.Writer.Flush()
		})
		if err != nil {
			a.logger.Error("Failed to stream greeting/farewell", zap.Error(err))
			a.writeErrorEvent(c, "Failed to stream greeting/farewell")
			return
		}

		// Save history
		postUUID, err := uuid.Parse(postID)
		if err != nil {
			a.logger.Error("Failed to parse post ID", zap.Error(err))
			return
		}

		// Save user question and AI response to history
		tokenCount := token.CountTokens(req.Prompt + fullText)
		if err := a.SaveChatHistory(c, &models.Post{ID: postUUID}, user, fullText, req.Prompt, tokenCount, os.Getenv("AI_MODEL")); err != nil {
			a.logger.Error("Failed to save chat history", zap.Error(err))
			return
		}
	}
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
		embedding32, err := a.llmClient.GenerateEmbedding(c.Request.Context(), phrase)
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
		return `คุณเป็นผู้ช่วยตอบคำถามจากบทความ ใช้เฉพาะข้อมูลในบทความเท่านั้น  
ห้ามเดา และห้ามใช้องค์ความรู้นอกเหนือจากนี้  

หากไม่พบคำตอบในบทความ ให้ตอบเพียง 2 บรรทัดเท่านั้นในรูปแบบนี้:
INTRODUCTION: <ข้อความแนะนำสั้น ๆ ว่าควรค้นหาอะไรเพิ่มเติม (อาจจะบอกว่าไม่พบเนื้อหาที่เกี่ยวข้องลองดูจากข้อมูลนี้ค่ะ)>
WEBSEARCH: <คำค้นหาที่ควรใช้>

ถ้าพบคำตอบในบทความ ให้ตอบตามปกติแบบสุภาพ กระชับ ไม่เกิน 5 ประโยค`
	}

	return fmt.Sprintf(`คุณเป็นผู้ช่วยตอบคำถามจากบทความด้านล่างนี้  
ใช้เฉพาะข้อมูลในบทความเท่านั้น ห้ามเดา และห้ามใช้องค์ความรู้นอกเหนือจากนี้  

หากไม่พบคำตอบในบทความ แต่เห็นว่าควรค้นต่อจากภายนอก ให้ตอบเพียง 2 บรรทัดเท่านั้นในรูปแบบนี้:
INTRODUCTION: <ข้อความแนะนำสั้น ๆ ว่าควรค้นหาอะไรเพิ่มเติม (อาจจะบอกว่าไม่พบเนื้อหาที่เกี่ยวข้องลองดูจากข้อมูลนี้ค่ะ)>
WEBSEARCH: <คำค้นหาที่ควรใช้>

ถ้าพบคำตอบในบทความ ให้ตอบตามปกติแบบสุภาพ กระชับ ไม่เกิน 5 ประโยค

เนื้อหา:
%s`, context)
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

	// Prepare messages for LLM
	messages := []llm_types.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: question},
	}

	fullText, err := a.sendLLMRequest(c.Request.Context(), messages, config, func(chunk string) {
		//  ไม่ต้อง stream ออกไปในที่นี้ เพราะเราจะ parse ทีหลัง
		c.Writer.Flush()
	})

	if err != nil {
		a.logger.Error("LLM error", zap.Error(err))
		a.writeErrorEvent(c, "LLM service error")
		return
	}

	postUUID, err := uuid.Parse(postID)
	if err != nil {
		a.writeErrorEvent(c, "Invalid post ID format")
		return
	}

	a.writeEvent(c, "start", "Streaming started")
	c.Writer.Flush()

	const webSearchPrefix = "WEBSEARCH:"
	const introductionPrefix = "INTRODUCTION:"

	// ตัวอย่างข้อความที่ LLM อาจตอบ:
	// INTRODUCTION:ขอโทษค่ะไม่พบคำตอบในบทความนี้คุณอาจลองค้นหาข้อมูลเพิ่มเติมเกี่ยวกับ"webserverคืออะไร"จากแหล่งข้อมูลอื่นค่ะWEBSEARCH:webserverคืออะไร

	a.logger.Debug("Full LLM response extracted",
		zap.String("response", fullText))

	// แปลงเป็น upper เพื่อค้น marker ได้ง่าย
	upperText := strings.ToUpper(fullText)
	webSearchIndex := strings.Index(upperText, webSearchPrefix)
	introIndex := strings.Index(upperText, introductionPrefix)

	if webSearchIndex != -1 {
		var introText, searchQuery string

		// ดึงส่วน introduction (ถ้ามี)
		if introIndex != -1 {
			start := introIndex + len(introductionPrefix)
			if webSearchIndex > start {
				introText = strings.TrimSpace(fullText[start:webSearchIndex])
			}
		}

		// ดึงส่วน search query
		searchQuery = strings.TrimSpace(fullText[webSearchIndex+len(webSearchPrefix):])

		if searchQuery == "" {
			a.writeErrorEvent(c, "Empty web search query")
			return
		}

		a.logger.Info("Agent requested web search",
			zap.String("query", searchQuery))

		//  เรียก external web search
		searchExternalResult, err := a.agentAgentToolWebSearchService.SearchExternalWeb(searchQuery)
		if err != nil {
			a.logger.Error("Web search failed", zap.Error(err))
			a.writeErrorEvent(c, "Web search failed")
			return
		}

		// --- stream ผลลัพธ์จาก external web ---
		combinedText := ""
		if introText != "" {
			combinedText = introText + "\n\n"
		}
		combinedText += searchExternalResult

		jsonResult, _ := json.Marshal(map[string]string{
			"text": combinedText,
		})
		fmt.Fprintf(c.Writer, "data: %s\n\n", jsonResult)
		c.Writer.Flush()

		a.writeEvent(c, "end", "done")

		// --- บันทึกประวัติการสนทนา ---
		combinedResponse := strings.TrimSpace(introText + "\n\n" + searchExternalResult)
		realTotalTokens := inputTokens + token.CountTokens(combinedResponse)
		if err := a.SaveChatHistory(c, &models.Post{ID: postUUID}, user, combinedResponse, question, realTotalTokens, config.Model); err != nil {
			a.logger.Error("Failed to save chat history", zap.Error(err))
		}
		return
	}

	// Use Gin's streaming mechanism
	c.Stream(func(w io.Writer) bool {
		jsonEncoded, _ := json.Marshal(map[string]string{"text": fullText})
		fmt.Fprintf(w, "data: %s\n\n", jsonEncoded)

		// Send the "end" event within the stream
		fmt.Fprintf(w, "event: end\ndata: done\n\n")
		return false // Close the stream after sending the full text and end event
	})

	// Token usage = input + streamed output
	realTotalTokens := inputTokens + token.CountTokens(fullText)

	// Save history
	if err := a.SaveChatHistory(c, &models.Post{ID: postUUID}, user, fullText, question, realTotalTokens, config.Model); err != nil {
		a.logger.Error("Failed to save chat history", zap.Error(err))
		// a.writeErrorEvent(c, "Failed to save chat history") // Cannot use c.Writer after c.Stream
	}
}

func (a *AIHandler) sendLLMRequest(ctx context.Context, messages []llm_types.ChatMessage, config RAGConfig, streamCallback func(string)) (string, error) {
	// Convert llm.ChatMessage to the format expected by the HTTP request for OpenRouter/Ollama
	httpMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		httpMessages[i] = map[string]string{"role": msg.Role, "content": msg.Content}
	}

	payload := map[string]interface{}{
		"model":    config.Model,
		"stream":   true,
		"messages": httpMessages,
	}

	body, _ := json.Marshal(payload)

	a.logger.Info("Sending request to LLM",
		zap.String("payload", string(body)))

	if config.UseSelfHost {
		a.logger.Info("Using self-hosted Ollama", zap.String("host", config.Host))
		resp, err := http.Post(config.Host+"/api/chat", "application/json", bytes.NewBuffer(body))
		if err != nil {
			return "", fmt.Errorf("ollama request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("ollama service error: %s", string(bodyBytes))
		}

		return a.streamHTTPResponse(resp, streamCallback)
	}

	// Using Bedrock via llmClient
	a.logger.Info("Using Bedrock via llmClient")
	return a.llmClient.StreamChatCompletion(ctx, messages, streamCallback)
}

func (a *AIHandler) streamHTTPResponse(resp *http.Response, streamCallback func(string)) (string, error) {
	reader := bufio.NewReader(resp.Body)
	var fullText strings.Builder

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				a.logger.Info("LLM stream finished")
			} else {
				a.logger.Error("Error reading LLM stream", zap.Error(err))
				return "", fmt.Errorf("error reading LLM stream: %w", err)
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
			break
		}

		content := a.parseStreamChunk(raw)
		if content != "" {
			fullText.WriteString(content)
			streamCallback(content)
		}
	}

	return fullText.String(), nil
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
func (a *AIHandler) SaveChatHistory(c *gin.Context, post *models.Post, user *models.User, responseText string, promt string, tokenUse int, modelName string) error {

	// Log all relevant parameters (post ID, response text, prompt, token usage, user email) before saving chat history for debugging and audit purposes
	a.logger.Info("Saving chat history",
		zap.String("post_id", post.ID.String()),
		zap.String("response_text", responseText),
		zap.String("prompt", promt),
		zap.Int("token_used", tokenUse),
		zap.String("user_email", user.Email))

	chat := &models.AIResponse{
		Response:  responseText,
		Prompt:    promt,
		UserID:    user.ID,
		PostID:    post.ID,
		TokenUsed: tokenUse,
		Success:   true,
		Model:     modelName,
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

func (h *AIHandler) GetChatsByPost(c *gin.Context) {
	postID := c.Param("post_id")
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}
	var userIDPtr *uuid.UUID
	if userObj, exists := c.Get("user"); exists {
		if user, ok := userObj.(*models.User); ok {
			userIDPtr = &user.ID
		}
	}
	chats, err := h.AIService.GetChatsByPost(postID, userIDPtr, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, chats)
}
