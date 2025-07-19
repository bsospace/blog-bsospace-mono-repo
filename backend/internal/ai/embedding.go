package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

func GetEmbedding(text string) ([]float32, error) {
	fmt.Println("---- GetEmbedding ----")

	ollamaURL := os.Getenv("AI_HOST")
	if ollamaURL == "" {
		return nil, fmt.Errorf("AI_HOST env var is empty")
	}
	fmt.Println("Embedding API:", ollamaURL)

	reqBody := EmbeddingRequest{Model: "nomic-embed-text", Prompt: text}
	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(ollamaURL+"/api/embeddings", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding API: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API returned non-200: %s", string(body))
	}

	var embeddingResp EmbeddingResponse
	err = json.Unmarshal(body, &embeddingResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %v", err)
	}

	embedding := embeddingResp.Embedding
	fmt.Println("Embedding length before padding:", len(embedding))

	const expectedDim = 384
	switch {
	case len(embedding) < expectedDim:
		padding := make([]float32, expectedDim-len(embedding))
		embedding = append(embedding, padding...)
	case len(embedding) > expectedDim:
		embedding = embedding[:expectedDim]
	}

	fmt.Println("Final embedding length:", len(embedding))
	return embedding, nil
}
