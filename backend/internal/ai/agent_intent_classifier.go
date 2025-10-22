package ai

import (
	"fmt"
	"os"
	"rag-searchbot-backend/internal/llm"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/post"
	"rag-searchbot-backend/pkg/utils"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// Intent represents the type of user question for routing agents in Agentic RAG.
type Intent string

// ScoredChunk represents a chunk of text with an associated score, used for ranking or relevance.
type ScoredChunk struct {
	Text  string  `json:"text"`
	Score float64 `json:"score"`
}

type RAGConfig struct {
	TopK                int     `json:"top_k"`
	SimilarityThreshold float64 `json:"similarity_threshold"` // ใช้ใน logic เดิม
	StrictThreshold     float64 `json:"strict_threshold"`     // สำหรับการกรองแบบเข้มงวด
}

const (
	DefaultTopK                = 10  // จำนวน context สูงสุดที่ดึงกลับมาต่อครั้ง (Top-K retrieval)
	DefaultSimilarityThreshold = 0.5 // ใช้เป็น threshold ปกติ ถ้าไม่มี override
	MaxTopK                    = 20  // กันไม่ให้ดึงเกินนี้แม้จะ fuzzy
	MinSimilarityThreshold     = 0.1 // ล่างสุดที่ยอมให้ผ่าน (fuzzy จริง ๆ)
	StrictThreshold            = 0.8 // ใช้ตอนต้องการ “เนื้อหาที่มั่นใจมาก” เท่านั้น
)

const (
	// blog_question: ผู้ใช้ถามเกี่ยวกับเนื้อหาภายในบทความที่กำลังเปิดอยู่
	IntentBlogQuestion Intent = "blog_question"

	// summarize_post: ผู้ใช้ต้องการให้สรุปบทความฉบับย่อ
	IntentSummarizePost Intent = "summarize_post"

	// search_blog: ผู้ใช้ต้องการให้ระบบค้นหาบทความที่เกี่ยวข้องกับคำถามนั้น (ไม่ใช่ถามบทความที่เปิดอยู่)
	IntentSearchBlog Intent = "search_blog"

	// generic_question: คำถามทั่วไปที่อาจไม่เกี่ยวข้องกับบทความปัจจุบัน
	// (สามารถใช้เป็น fallback หรือนำไปค้นต่อ)
	IntentGeneric Intent = "generic_question"

	// external_question: คำถามเกี่ยวกับความรู้ทั่วไปภายนอก blog เช่น API, ความรู้เฉพาะทาง, ข่าว ฯลฯ
	// ต้องใช้ ExternalKnowledgeAgent ในการค้นหาหรือสรุปจากแหล่งภายนอก (Wikipedia, API, etc.)
	IntentExternalQuestion Intent = "external_question"

	// greeting_farewell: การทักทายหรือกล่าวลา เช่น "สวัสดี", "ขอบคุณนะ", "บาย"
	IntentGreetingFarewell Intent = "greeting_farewell"

	// unknown: ไม่สามารถระบุเจตนาหรือไม่เข้าใจคำถามของผู้ใช้
	IntentUnknown Intent = "unknown"
)

type AgentIntentClassifierServiceInterface interface {
	// ClassifyIntent รับข้อความและคืนค่าเจตนาของผู้ใช้
	ClassifyIntent(message string, post *models.Post) (string, error)
}

type agentIntentClassifierService struct {
	logger    *zap.Logger
	PosRepo   post.PostRepositoryInterface
	LLmClient llm.LLM
}

func NewAgentIntentClassifier(logger *zap.Logger, posRepo post.PostRepositoryInterface, LLmClient llm.LLM) AgentIntentClassifierServiceInterface {
	return &agentIntentClassifierService{
		logger:    logger,
		PosRepo:   posRepo,
		LLmClient: LLmClient,
	}
}

func (a *agentIntentClassifierService) ClassifyIntent(message string, post *models.Post) (string, error) {

	a.logger.Debug("Classifying intent starttttt", zap.String("message", message))
	a.logger.Debug("Post ID for intent classification", zap.String("post_id", post.ID.String()),
		zap.String("post_title", post.Title))

	classify, err := a.RetrieveContext(message, post)
	if err != nil {
		return "", err
	}

	return classify, err
}

// Retrived context from blog content
func (a *agentIntentClassifierService) RetrieveContext(message string, post *models.Post) (string, error) {

	// log
	a.logger.Debug("Retrieving context for message", zap.String("message", message))
	// log post ID
	a.logger.Debug("Post ID for context retrieval", zap.String("post_id", post.ID.String()))

	// allScoredChunks := []ScoredChunk{}
	phrases := SplitQuestionToPhrases(message)

	// logs
	a.logger.Debug("Split question into phrases", zap.Strings("phrases", phrases))

	var allScoredChunks []ScoredChunk
	// log all phrases for debugging
	for _, phrase := range phrases {
		embedding32, err := a.LLmClient.GenerateEmbedding(nil, phrase)
		a.logger.Debug("Embedding for phrase", zap.String("phrase", phrase), zap.Any("embedding", len(embedding32)))
		if err != nil {
			a.logger.Warn("Failed to get embedding for phrase", zap.String("phrase", phrase), zap.Error(err))
			continue
		}

		embedding := make([]float64, len(embedding32))
		for i, v := range embedding32 {
			embedding[i] = float64(v)
		}

		// log scored chunks for phrase
		a.logger.Debug("Retrieving scored chunks for phrase", zap.String("phrase", phrase))

		scoredChunks, err := a.retrieveAndScoreChunks(post.ID.String(), embedding)
		if err != nil {
			a.logger.Warn("Failed to retrieve and score chunks", zap.String("post_id", post.ID.String()), zap.Error(err))
			continue
		}

		a.logger.Debug("Scored chunks retrieved", zap.Int("count", len(scoredChunks)))

		allScoredChunks = append(allScoredChunks, scoredChunks...)
	}

	// log total scored chunks
	a.logger.Debug("Total scored chunks after all phrases", zap.Int("total_count", len(allScoredChunks)))

	// select top chunks based on config
	config := RAGConfig{
		TopK:                DefaultTopK,
		SimilarityThreshold: DefaultSimilarityThreshold,
		StrictThreshold:     StrictThreshold,
	}

	selectedChunks := a.selectTopChunks(allScoredChunks, config)

	// log selected chunks
	a.logger.Debug("Selected chunks after filtering",
		zap.Int("count", len(selectedChunks)),
		zap.Float64("strict_threshold", config.StrictThreshold),
		zap.Float64("similarity_threshold", config.SimilarityThreshold))

	// log chunks content
	for i, chunk := range selectedChunks {
		a.logger.Debug("Selected chunk",
			zap.Int("index", i),
			zap.Float64("score", chunk.Score),
			zap.String("preview", a.truncateText(chunk.Text, 50)),
		)
	}

	if len(selectedChunks) == 0 {
		intent, err := a.ClassifyWithOpenRouter(message, []string{post.Title + "\n\n" + post.Description})
		if err != nil {
			a.logger.Warn("Failed to classify intent with no context", zap.Error(err))
			return string(IntentUnknown), err
		}
		a.logger.Debug("Intent classified with no context", zap.String("intent", string(intent)))
		return string(intent), nil
	}

	// Join selected chunks into a single context string
	var contextBuilder strings.Builder
	for i, chunk := range selectedChunks {
		if i > 0 {
			contextBuilder.WriteString("\n\n") // Separate chunks with double newlines
		}
		contextBuilder.WriteString(chunk.Text)
	}

	// send context to OpenRouter for classification
	context := contextBuilder.String()
	a.logger.Debug("Final context for classification", zap.String("context", context))

	intent, err := a.ClassifyWithOpenRouter(message, []string{context})
	if err != nil {
		a.logger.Warn("Failed to classify intent", zap.Error(err))
		return "", err
	}

	a.logger.Debug("Intent classified", zap.String("intent", string(intent)))

	return string(intent), nil
}

func (a *agentIntentClassifierService) retrieveAndScoreChunks(postID string, questionEmbedding []float64) ([]ScoredChunk, error) {
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

		// a.logger.Debug("Chunk scored",
		// 	zap.Float64("score", score),
		// 	zap.String("preview", a.truncateText(chunk.Content, 50)))

	}

	// Sort by score descending
	sort.Slice(scoredChunks, func(i, j int) bool {
		return scoredChunks[i].Score > scoredChunks[j].Score
	})

	return scoredChunks, nil
}

// call open router for classification
func (a *agentIntentClassifierService) ClassifyWithOpenRouter(message string, context []string) (string, error) {
	// Join context chunks (จำกัดจำนวนหรือความยาวตามความเหมาะสม)
	joinedContext := strings.Join(context, "\n\n")
	if len(joinedContext) > 4000 {
		joinedContext = joinedContext[:4000] // ตัดความยาวกัน LLM token overflow
	}

	systemPrompt := `
		You are an intent classifier for a blog Q&A system.
		Your job is to classify the user's question based on the given blog context.

		Here are the possible intents:
		- blog_question: ถามเนื้อหาบทความ เช่น "บทความนี้เกี่ยวกับอะไร", "RAG คืออะไร"
		- summarize_post: ขอให้สรุปบทความ เช่น "ช่วยสรุปให้หน่อย"
		- greeting_farewell: ทักทายหรือกล่าวลา เช่น "สวัสดี", "ลาก่อน"

		Use the blog context to help guide your classification.

		Example answers: blog_question

		Blog Context:
		"""` + joinedContext + `"""`

	payload := map[string]interface{}{
		"model":  os.Getenv("AI_MODEL"),
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": message},
		},
	}

	config := map[string]string{
		"api_key":       os.Getenv("AI_API_KEY"),
		"model":         os.Getenv("AI_MODEL"),
		"host":          os.Getenv("AI_HOST"),
		"use_self_host": os.Getenv("AI_SELF_HOST"),
	}

	a.logger.Debug("Sending intent classification request to OpenRouter",
		zap.String("message", message),
		zap.Int("context_len", len(context)),
	)

	resp, err := SendLLMRequestToOpenRouter(payload, config)
	if err != nil {
		a.logger.Warn("Failed to classify intent via OpenRouter", zap.Error(err))
		return string(IntentUnknown), err
	}

	intent := parseIntentFromLLM(resp)

	// log raw intent for debugging
	a.logger.Debug("Raw intent from OpenRouter response", zap.String("raw_intent", resp))

	a.logger.Debug("Intent classified by OpenRouter", zap.String("intent", string(intent)))

	return string(intent), nil
}

func (a *agentIntentClassifierService) truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
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

func (a *agentIntentClassifierService) selectTopChunks(scoredChunks []ScoredChunk, config RAGConfig) []ScoredChunk {
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

	// ถ้าไม่มีเลย ไม่ต้องเอา เพราะเป็น classification agent
	if len(selectedChunks) == 0 {
		return nil
	}

	return selectedChunks
}
