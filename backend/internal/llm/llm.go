package llm

import "context"

type LLM interface {
	InvokeLLM(ctx context.Context, prompt string) (string, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}
