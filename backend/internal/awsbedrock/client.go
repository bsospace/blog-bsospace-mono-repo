package awsbedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"rag-searchbot-backend/config"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type BedrockClient struct {
	client *bedrockruntime.Client
	cfg    config.Config
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

//
// ────────────────────────────────
//   TEXT GENERATION (LLM)
// ────────────────────────────────
//

func (bc *BedrockClient) InvokeLLM(ctx context.Context, prompt string) (string, error) {
	modelID := bc.cfg.AWSBedrockLLMModel
	if modelID == "" {
		modelID = "amazon.titan-text-express-v1"
	}

	var requestBody map[string]interface{}

	switch {
	//  Titan Text
	case strings.HasPrefix(modelID, "amazon.titan-text"):
		requestBody = map[string]interface{}{
			"prompt": prompt,
		}

	//  Meta Llama (Llama 3.x)
	case strings.Contains(modelID, "meta.llama3"):
		requestBody = map[string]interface{}{
			"prompt":      prompt,
			"max_gen_len": 512,
			"temperature": 0.7,
			"top_p":       0.9,
		}

	//  Anthropic Claude 3
	case strings.Contains(modelID, "anthropic.claude"):
		requestBody = map[string]interface{}{
			"messages": []map[string]string{
				{"role": "user", "content": prompt},
			},
		}

	//  fallback
	default:
		requestBody = map[string]interface{}{
			"prompt": prompt,
		}
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

	//  Handle Llama-style output
	if generation, ok := responseBody["generation"].(string); ok {
		return generation, nil
	}
	if results, ok := responseBody["results"].([]interface{}); ok && len(results) > 0 {
		if result, ok := results[0].(map[string]interface{}); ok {
			if text, ok := result["outputText"].(string); ok {
				return text, nil
			}
		}
	}
	if completion, ok := responseBody["completion"].(string); ok {
		return completion, nil
	}

	return "", fmt.Errorf("could not extract generated text from Bedrock response")
}

//
// ────────────────────────────────
//   EMBEDDING GENERATION
// ────────────────────────────────
//

func (bc *BedrockClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	modelID := bc.cfg.AWSBedrockEmbeddingModel
	if modelID == "" {
		modelID = "amazon.titan-embed-text-v1"
	}

	//  Titan Embeddings ใช้ key "inputText"
	requestBody := map[string]string{
		"inputText": text,
	}

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
