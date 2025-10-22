package awsbedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"rag-searchbot-backend/config"
	"rag-searchbot-backend/internal/llm_types"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

type BedrockClient struct {
	client *bedrockruntime.Client
	cfg    config.Config
}

// ────────────────────────────────
// Helper structs
// ────────────────────────────────

type claudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeMessage struct {
	Role    string               `json:"role"`
	Content []claudeContentBlock `json:"content"`
}

type claudeRequestBody struct {
	Messages         []claudeMessage `json:"messages"`
	AnthropicVersion string          `json:"anthropic_version"`
	MaxTokens        int             `json:"max_tokens"`
	Temperature      float64         `json:"temperature"`
	TopP             float64         `json:"top_p"`
}

func NewBedrockClient(cfg config.Config) (*BedrockClient, error) {
	var awsCfg aws.Config
	var err error

	if cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" {
		awsCfg, err = awsConfig.LoadDefaultConfig(
			context.TODO(),
			awsConfig.WithRegion(cfg.AWSRegion),
			awsConfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, ""),
			),
		)
	} else {
		awsCfg, err = awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(cfg.AWSRegion))
	}

	if err != nil {
		log.Printf("Error loading AWS config: %v", err)
		return nil, err
	}

	return &BedrockClient{
		client: bedrockruntime.NewFromConfig(awsCfg),
		cfg:    cfg,
	}, nil
}

// ────────────────────────────────
// TEXT GENERATION (Non-Streaming)
// ────────────────────────────────

func (bc *BedrockClient) InvokeLLM(ctx context.Context, prompt string) (string, error) {
	modelID := bc.cfg.AWSBedrockLLMModel
	if modelID == "" {
		modelID = "amazon.titan-text-express-v1"
	}

	var requestBody map[string]interface{}

	switch {
	case strings.HasPrefix(modelID, "amazon.titan-text"):
		requestBody = map[string]interface{}{
			"prompt": prompt,
		}

	case strings.Contains(modelID, "meta.llama3"):
		// ใช้ prompt format ของ Meta Llama 3
		formattedPrompt := fmt.Sprintf(`
		<|begin_of_text|><|start_header_id|>user<|end_header_id|>
		%s
		<|eot_id|>
		<|start_header_id|>assistant<|end_header_id|>
`, prompt)

		requestBody = map[string]interface{}{
			"prompt":      formattedPrompt,
			"max_gen_len": 512,
			"temperature": 0.7,
			"top_p":       0.9,
		}

	case strings.Contains(modelID, "anthropic.claude"):
		requestBody = map[string]interface{}{
			"messages": []map[string]string{
				{"role": "user", "content": prompt},
			},
			"anthropic_version": "bedrock-2023-05-31",
			"temperature":       0.7,
			"top_p":             0.9,
		}

	default:
		requestBody = map[string]interface{}{"prompt": prompt}
	}

	marshalledBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	output, err := bc.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		ContentType: aws.String("application/json"),
		Body:        marshalledBody,
	})
	if err != nil {
		return "", fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}

	var responseBody map[string]interface{}
	if err := json.Unmarshal(output.Body, &responseBody); err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	if generation, ok := responseBody["generation"].(string); ok {
		return generation, nil
	}
	if completion, ok := responseBody["completion"].(string); ok {
		return completion, nil
	}

	var results string
	if results, ok := responseBody["results"].([]interface{}); ok && len(results) > 0 {
		if result, ok := results[0].(map[string]interface{}); ok {
			if text, ok := result["outputText"].(string); ok {
				return text, nil
			}
		}
	}

	// if have ctx make stream

	return results, fmt.Errorf("could not extract generated text from Bedrock response")
}

// ────────────────────────────────
// STREAMING CHAT COMPLETION
// ────────────────────────────────

func (bc *BedrockClient) StreamChatCompletion(
	ctx context.Context,
	messages []llm_types.ChatMessage,
	streamCallback func(string),
) (string, error) {

	modelID := bc.cfg.AWSBedrockLLMModel
	if modelID == "" {
		modelID = "anthropic.claude-3-sonnet-20240229-v1:0"
	}

	var marshalledBody []byte
	var err error

	switch {
	//  Anthropic Claude 3
	case strings.Contains(modelID, "anthropic.claude"):
		bedrockMessages := make([]claudeMessage, len(messages))
		for i, msg := range messages {
			bedrockMessages[i] = claudeMessage{
				Role: string(msg.Role),
				Content: []claudeContentBlock{
					{Type: "text", Text: msg.Content},
				},
			}
		}

		requestBody := claudeRequestBody{
			Messages:         bedrockMessages,
			AnthropicVersion: "bedrock-2023-05-31",
			MaxTokens:        4096,
			Temperature:      0.7,
			TopP:             0.9,
		}
		marshalledBody, err = json.Marshal(requestBody)

	//  Meta Llama 3.x
	case strings.Contains(modelID, "meta.llama3"):
		var promptBuilder strings.Builder
		for _, msg := range messages {
			promptBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
		requestBody := map[string]interface{}{
			"prompt":      promptBuilder.String(),
			"max_gen_len": 4096,
			"temperature": 0.7,
			"top_p":       0.9,
		}
		marshalledBody, err = json.Marshal(requestBody)

	//  Titan / fallback
	default:
		var promptBuilder strings.Builder
		for _, msg := range messages {
			promptBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
		requestBody := map[string]interface{}{
			"prompt":      promptBuilder.String(),
			"max_gen_len": 4096,
			"temperature": 0.7,
			"top_p":       0.9,
		}
		marshalledBody, err = json.Marshal(requestBody)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	output, err := bc.client.InvokeModelWithResponseStream(ctx, &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(modelID),
		ContentType: aws.String("application/json"),
		Body:        marshalledBody,
	})
	if err != nil {
		return "", fmt.Errorf("failed to invoke Bedrock model with stream: %w", err)
	}
	defer output.GetStream().Close()

	var fullResponse strings.Builder
	// ใน loop ที่อ่าน event stream
	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			var responseBody map[string]interface{}
			if err := json.Unmarshal(v.Value.Bytes, &responseBody); err != nil {
				return "", fmt.Errorf("failed to unmarshal stream chunk: %w", err)
			}

			var chunkText string
			if completion, ok := responseBody["completion"].(string); ok {
				chunkText = completion
			} else if generation, ok := responseBody["generation"].(string); ok {
				chunkText = generation
			} else if generatedText, ok := responseBody["generated_text"].(string); ok {
				chunkText = generatedText
			}

			//  ตัด prefix "assistant:" ออก (กรณี model ส่ง prefix มาด้วย)
			cleaned := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(chunkText, "assistant:"), "Assistant:"))

			if cleaned != "" {
				streamCallback(cleaned)
				fullResponse.WriteString(cleaned)
			}

		default:
			log.Printf("received unknown event type: %T", v)
		}
	}

	return fullResponse.String(), nil
}

// ────────────────────────────────
// EMBEDDING GENERATION
// ────────────────────────────────

func (bc *BedrockClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {

	if ctx == nil {
		ctx = context.Background()
	}
	modelID := bc.cfg.AWSBedrockEmbeddingModel
	if modelID == "" {
		modelID = "amazon.titan-embed-text-v1"
	}

	requestBody := map[string]string{"inputText": text}

	marshalledBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	output, err := bc.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		ContentType: aws.String("application/json"),
		Body:        marshalledBody,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to invoke Bedrock embedding model: %w", err)
	}

	var responseBody map[string]interface{}
	if err := json.Unmarshal(output.Body, &responseBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	embedding, ok := responseBody["embedding"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no embedding found in Bedrock response")
	}

	floatEmbedding := make([]float32, len(embedding))
	for i, v := range embedding {
		if f, ok := v.(float64); ok {
			floatEmbedding[i] = float32(f)
		} else {
			return nil, fmt.Errorf("invalid embedding format")
		}
	}

	return floatEmbedding, nil
}
