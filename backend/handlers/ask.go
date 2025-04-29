package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"rag-searchbot-backend/internal/ollama"
	"rag-searchbot-backend/internal/storage"
	"rag-searchbot-backend/utils"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

type AskRequest struct {
	Question string `json:"question"`
}

func AskHandler(c *gin.Context) {
	var req AskRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 1. Embed the question
	questionEmbedding, err := ollama.GetEmbedding(req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to embed question"})
		return
	}

	// 2. Find Top 3 best matching chunks
	type ScoredChunk struct {
		Text  string
		Score float64
	}

	var scoredChunks []ScoredChunk
	for _, chunk := range storage.GetChunks() {
		score := utils.CosineSimilarity(chunk.Embedding, questionEmbedding)
		scoredChunks = append(scoredChunks, ScoredChunk{
			Text:  chunk.Text,
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
		c.JSON(http.StatusNotFound, gin.H{"error": "No relevant context found"})
		return
	}

	// 3. Call LLM with fullContext
	answer, err := callOllamaChat(fullContext, req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call LLM", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"answer": answer,
	})
}

type ChatRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

func callOllamaChat(context, question string) (string, error) {
	prompt := `
You are a helpful assistant.

Use the following context to answer the question as accurately as possible.
If the answer is unclear or incomplete, politely say so, but try to help.

Context:
` + context + `

Question: ` + question + `

Answer:
`

	reqBody := ChatRequest{
		Model:  "scb10x/llama3.1-typhoon2-8b-instruct:latest",
		Prompt: prompt,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	ollamaURL := os.Getenv("OLLAMA_HOST")
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
