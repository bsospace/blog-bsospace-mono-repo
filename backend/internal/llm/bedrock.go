package llm

import (
	"context"
	"rag-searchbot-backend/internal/awsbedrock"
	"rag-searchbot-backend/internal/llm_types"
)

type BedrockLLM struct {
	client *awsbedrock.BedrockClient
}

func NewBedrockLLM(client *awsbedrock.BedrockClient) *BedrockLLM {
	return &BedrockLLM{
		client: client,
	}
}

func (b *BedrockLLM) InvokeLLM(ctx context.Context, prompt string) (string, error) {
	return b.client.InvokeLLM(ctx, prompt)
}

func (b *BedrockLLM) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return b.client.GenerateEmbedding(ctx, text)
}

func (b *BedrockLLM) StreamChatCompletion(ctx context.Context, messages []llm_types.ChatMessage, streamCallback func(string)) (string, error) {
	return b.client.StreamChatCompletion(ctx, messages, streamCallback)
}
