package llm

import (
	"context"

	"rag-searchbot-backend/internal/llm_types"
)

type LLM interface {
	InvokeLLM(ctx context.Context, prompt string) (string, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	StreamChatCompletion(ctx context.Context, messages []llm_types.ChatMessage, streamCallback func(string)) (string, error)
}
