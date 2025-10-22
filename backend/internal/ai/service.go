package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"rag-searchbot-backend/internal/llm"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/post"
	"rag-searchbot-backend/pkg/utils"
	"sort"
	"strings"

	"github.com/google/uuid"
)

type AIService struct {
	PosRepo                 post.PostRepositoryInterface
	TaskEnqueuer            *TaskEnqueuer
	AIRepo                  AIRepositoryInterface
	IntentClassifierService AgentIntentClassifierServiceInterface
	llmClient               llm.LLM
}

func NewAIService(posRepo post.PostRepositoryInterface, enqueuer *TaskEnqueuer, aiRepo AIRepositoryInterface, intentClassifierService AgentIntentClassifierServiceInterface, llmClient llm.LLM) *AIService {
	return &AIService{
		PosRepo:                 posRepo,
		TaskEnqueuer:            enqueuer,
		AIRepo:                  aiRepo,
		IntentClassifierService: intentClassifierService,
		llmClient:               llmClient,
	}
}

func (s *AIService) OpenAIMode(postID string, userData *models.User) (bool, error) {
	post, err := s.PosRepo.GetByID(postID)
	if err != nil {
		return false, err
	}
	if post == nil {
		return false, nil
	}

	if post.AIChatOpen {
		return false, nil // AI chat already open
	}

	if post.AuthorID != userData.ID {
		return false, nil // User is not the author of the post
	}

	_, err = s.TaskEnqueuer.EnqueuePostEmbedding(post, userData)
	if err != nil {
		return false, err
	}

	return true, nil
}

// DisableOpenAIMode disables the OpenAI mode for a post
func (s *AIService) DisableOpenAIMode(postID string, userData *models.User) (bool, error) {
	post, err := s.PosRepo.GetByID(postID)
	if err != nil {
		return false, err
	}
	if post == nil {
		return false, nil
	}

	if !post.AIChatOpen {
		return false, nil // AI chat already disabled
	}

	if post.AuthorID != userData.ID {
		return false, nil // User is not the author of the post
	}

	post.AIChatOpen = false
	post.AIReady = false
	err = s.PosRepo.Update(post)
	if err != nil {
		return false, err
	}

	// delete all embeddings for this post
	if err := s.PosRepo.DeleteEmbeddingsByPostID(postID); err != nil {
		return false, err
	}

	return true, nil
}

type AskRequest struct {
	Question string `json:"question"`
}

func (s *AIService) ChatStream(postID string, userData *models.User, prompt string, onChunk func(string)) error {
	var req AskRequest
	if err := json.Unmarshal([]byte(prompt), &req); err != nil {
		return err
	}

	post, err := s.PosRepo.GetByID(postID)
	if err != nil || post == nil || !post.AIChatOpen {
		return fmt.Errorf("post not found or AI not enabled")
	}

	questionEmbedding, err := GetEmbedding(req.Question)
	if err != nil {
		return err
	}

	chunks, err := s.PosRepo.GetEmbeddingByPostID(postID)
	if err != nil {
		return err
	}

	type ScoredChunk struct {
		Text  string
		Score float64
	}

	var scoredChunks []ScoredChunk
	for _, chunk := range chunks {
		score := utils.CosineSimilarity(chunk.Vector.Slice(), questionEmbedding)
		scoredChunks = append(scoredChunks, ScoredChunk{
			Text:  chunk.Content,
			Score: score,
		})
	}

	sort.Slice(scoredChunks, func(i, j int) bool {
		return scoredChunks[i].Score > scoredChunks[j].Score
	})

	topChunks := []string{}
	for i := 0; i < 3 && i < len(scoredChunks); i++ {
		topChunks = append(topChunks, scoredChunks[i].Text)
	}

	fullContext := strings.Join(topChunks, "\n\n")
	if fullContext == "" {
		fullContext = "There is no relevant information from the document. Answer the question as best as you can or inform the user you cannot answer."
	}

	// Call streaming LLM
	return StreamAIResponse(fullContext, req.Question, onChunk)
}

func StreamAIResponse(context, question string, onChunk func(string)) error {
	models := os.Getenv("AI_MODEL")
	payload := map[string]interface{}{
		"model":  models,
		"stream": true,
		"messages": []map[string]string{
			{"role": "system", "content": context},
			{"role": "user", "content": question},
		},
	}

	body, _ := json.Marshal(payload)
	ollamaURL := os.Getenv("AI_HOST")
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	fmt.Println("Response status:", resp.StatusCode)
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')

		fmt.Println("Received line:", string(line))
		if err != nil {
			break
		}

		if bytes.HasPrefix(line, []byte("data: ")) {
			raw := bytes.TrimSpace(line[6:])

			if len(raw) == 0 || bytes.Equal(raw, []byte("[DONE]")) {
				continue
			}

			var chunk map[string]interface{}
			if err := json.Unmarshal(raw, &chunk); err != nil {
				continue
			}

			if message, ok := chunk["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					// ส่งแบบ JSON ที่ฝั่ง client รับง่าย
					jsonEncoded, _ := json.Marshal(map[string]string{"text": content})
					onChunk(string(jsonEncoded))
				}
			}
		}
	}
	return nil
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"` // false = return full output
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type ChatRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

func (s *AIService) CreateChat(chat *models.AIResponse, postID string, user *models.User) error {

	if chat == nil {
		return fmt.Errorf("chat cannot be nil")
	}

	if chat.UserID == uuid.Nil || chat.PostID == uuid.Nil {
		return fmt.Errorf("chat must have UserID and PostID")
	}

	if chat.Prompt == "" || chat.Response == "" {
		return fmt.Errorf("chat must have Prompt and Response")
	}

	return s.AIRepo.CreateChat(chat)
}

// classifyContent classifies the content of a post using AI
func (s *AIService) ClassifyContent(message string) (string, error) {
	if message == "" {
		return "", fmt.Errorf("message cannot be empty")
	}

	// Use the injected IntentClassifierService
	messageType, err := s.IntentClassifierService.ClassifyIntent(message, nil)
	if err != nil {
		return "", fmt.Errorf("failed to classify intent: %w", err)
	}

	if Intent(messageType) == IntentUnknown {
		return "", fmt.Errorf("unknown message type: %s", messageType)
	}

	return messageType, nil
}

func ToChatDTO(chat models.AIResponse) ChatDTO {
	return ChatDTO{
		ID:        chat.ID,
		UsedAt:    chat.UsedAt,
		Prompt:    chat.Prompt,
		Response:  chat.Response,
		TokenUsed: chat.TokenUsed,
		Success:   chat.Success,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
	}
}

func (s *AIService) GetChatsByPost(postID string, userID *uuid.UUID, limit, offset int) ([]ChatDTO, error) {
	chats, err := s.AIRepo.GetChatsByPost(postID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	var dtos []ChatDTO
	for _, chat := range chats {
		dtos = append(dtos, ToChatDTO(chat))
	}
	return dtos, nil
}

func (s *AIService) StreamPostSummary(ctx context.Context, prompt, plaintextContent string, onChunk func(string)) (string, error) {
	systemPrompt := `คุณคือผู้ช่วยสรุปบทความ จงสรุปบทความที่ให้มาอย่างกระชับและได้ใจความสำคัญ ไม่เกิน 5 ประโยค`
	fullPrompt := fmt.Sprintf("%s\nบทความ: %s\nคำสั่ง: %s", systemPrompt, plaintextContent, prompt)

	resp, err := s.llmClient.InvokeLLM(ctx, fullPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to invoke LLM for post summary: %w", err)
	}

	onChunk(resp)
	return resp, nil
}

func (s *AIService) StreamGreetingFarewell(ctx context.Context, prompt string, onChunk func(string)) (string, error) {
	systemPrompt := `คุณคือผู้ช่วยตอบคำทักทายและกล่าวลา จงตอบกลับอย่างสุภาพและเป็นมิตรจากข้อความนี้ และตอบกลับมาในรูปแบบสั้นๆ`
	fullPrompt := fmt.Sprintf("%s\n%s", systemPrompt, prompt)

	resp, err := s.llmClient.InvokeLLM(ctx, fullPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to invoke LLM for greeting/farewell: %w", err)
	}

	onChunk(resp)
	return resp, nil
}
