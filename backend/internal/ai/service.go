package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/post"
	"rag-searchbot-backend/pkg/utils"
	"sort"
	"strings"
)

type AIService struct {
	PosRepo      post.PostRepositoryInterface
	TaskEnqueuer *TaskEnqueuer
}

func NewAIService(posRepo post.PostRepositoryInterface, enqueuer *TaskEnqueuer) *AIService {
	return &AIService{
		PosRepo:      posRepo,
		TaskEnqueuer: enqueuer,
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
	err = s.PosRepo.Update(post)
	if err != nil {
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

func callAIChat(context, question string) (string, error) {
	prompt := `
You are a helpful assistant.

Use the following context to answer the question as accurately as possible.
If the answer is unclear or incomplete, politely say so, but try to help.

Context:
` + context + `

Question: ` + question + `

Answer:
`

	llmModel := os.Getenv("AI_MODEL")
	reqBody := ChatRequest{
		Model:  llmModel,
		Prompt: prompt,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	ollamaURL := os.Getenv("AI_HOST")
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read streaming response
	scanner := bufio.NewScanner(resp.Body)
	var finalResponse string

	for scanner.Scan() {
		line := scanner.Bytes()
		var partial ChatResponse
		if err := json.Unmarshal(line, &partial); err != nil {
			continue
		}
		finalResponse += partial.Response
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return finalResponse, nil
}
