package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/post"
	"strings"

	"github.com/gin-gonic/gin"
)

// Intent represents the type of user question.

// Agent is the interface for all agents.
type Agent interface {
	Handle(question string, ctx *AgentContext) (string, error)
}

// AgentContext provides context for agent processing.
type AgentContext struct {
	Post    *models.Post
	User    *models.User
	PosRepo post.PostRepositoryInterface
	// เพิ่ม field อื่นๆ ได้ เช่น Embeddings, Logger, etc.
}

// ErrInsufficientContext is returned by BlogAgent when RAG context is not enough.
var ErrInsufficientContext = fmt.Errorf("insufficient context for RAG")

// IntentClassifier uses LLM (OpenRouter) to classify the question intent.
func IntentClassifier(question string) Intent {
	// Prepare system prompt for LLM
	systemPrompt := `
You are an intent classifier for a blog Q&A system. 
Classify the user's question into one of these intents:
- blog_question: ถามเนื้อหาบทความ เช่น "บทความนี้เกี่ยวกับอะไร", "RAG คืออะไร"
- summarize_post: ขอให้สรุปบทความ เช่น "ช่วยสรุปให้หน่อย"
- greeting_farewell: ทักทายหรือกล่าวลา เช่น "สวัสดี", "ลาก่อน"
- unknown: ไม่เข้าใจคำถาม`

	payload := map[string]interface{}{
		"model":  os.Getenv("AI_MODEL"),
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": question},
		},
	}
	config := map[string]string{
		"api_key":       os.Getenv("AI_API_KEY"),
		"model":         os.Getenv("AI_MODEL"),
		"host":          os.Getenv("AI_HOST"),
		"use_self_host": os.Getenv("AI_SELF_HOST"),
	}
	log.Printf("IntentClassifier payload: %+v", payload)
	resp, err := SendLLMRequestToOpenRouter(payload, config)
	if err != nil {
		log.Printf("IntentClassifier error: %v", err)
		return IntentUnknown
	}

	// log.Printf("IntentClassifier LLM raw response: %q", resp)
	return parseIntentFromLLM(resp)
}

// Helper methods
func writeEvent(c *gin.Context, event, data string) {
	fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data)
	c.Writer.Flush()
}

func writeErrorEvent(c *gin.Context, message string) {
	fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", message)
	c.Writer.Flush()
}

func parseStreamChunk(raw []byte) string {
	var chunk map[string]interface{}
	if err := json.Unmarshal(raw, &chunk); err != nil {
		return ""
	}

	// Handle Ollama format
	if message, ok := chunk["message"].(map[string]interface{}); ok {
		if content, ok := message["content"].(string); ok {
			return content
		}
	}

	// Handle OpenAI/OpenRouter format
	if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
		choice := choices[0].(map[string]interface{})
		if delta, ok := choice["delta"].(map[string]interface{}); ok {
			if content, ok := delta["content"].(string); ok {
				return content
			}
		}
	}

	return ""
}

// AGENT : Summarize post agent
func StreamPostSummaryAgent(c *gin.Context, question string, htmlContent string) string {
	// 1. สร้าง system prompt
	systemPrompt := fmt.Sprintf(`คุณเป็นผู้ช่วยสรุปเนื้อหาบทความใน BSO Space Blog
โปรดสรุปเนื้อหาด้านล่างนี้ให้กระชับ เข้าใจง่าย และเป็นมิตร
-----
%s
-----`, htmlContent)

	// 2. คำนวณ token
	// inputText := systemPrompt + "\n" + question
	// inputTokens := token.CountTokens(inputText)

	// 3. เตรียม payload
	payload := map[string]interface{}{
		"model":  os.Getenv("AI_MODEL"),
		"stream": true,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": question},
		},
	}

	// 4. เตรียม HTTP request
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AI_API_KEY"))
	req.Header.Set("HTTP-Referer", "https://blog.bsospace.com")
	req.Header.Set("X-Title", "https://blog.bsospace.com")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		writeErrorEvent(c, "เรียกใช้ LLM ไม่สำเร็จ")
		return ""
	}
	defer resp.Body.Close()

	// 5. ตั้ง header สำหรับ SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Writer.Flush()

	// 6. ส่ง event เริ่ม
	writeEvent(c, "start", "เริ่มการสรุปเนื้อหา")

	// 7. อ่านและส่ง stream
	fullText := streamLLMResponse(c, resp)

	// 8. ส่ง event จบ
	writeEvent(c, "end", "สรุปเนื้อหาเสร็จสิ้น")

	// 9. return full text
	return fullText
}

// AGENT: greeting_farewell agent
func StreamGreetingFarewellAgent(c *gin.Context, question string) string {
	// 1. สร้าง system prompt
	systemPrompt := `คุณเป็นผู้ช่วยทักทายและกล่าวลาใน BSO Space
โปรดตอบคำถามทักทายหรือกล่าวลาอย่างเป็นมิตรและสุภาพ 
คำถาม: ` + question

	// 2. เตรียม payload
	payload := map[string]interface{}{
		"model":  os.Getenv("AI_MODEL"),
		"stream": true,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": question},
		},
	}

	// 3. เตรียม HTTP request
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AI_API_KEY"))
	req.Header.Set("HTTP-Referer", "https://blog.bsospace.com")
	req.Header.Set("X-Title", "https://blog.bsospace.com")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		writeErrorEvent(c, "เรียกใช้ LLM ไม่สำเร็จ")
		return ""
	}
	defer resp.Body.Close()

	// 4. ตั้ง header สำหรับ SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Writer.Flush()

	// 5. ส่ง event เริ่ม
	writeEvent(c, "start", "เริ่มการตอบคำถามทักทาย/กล่าวลา")

	// 6. อ่านและส่ง stream
	fullText := streamLLMResponse(c, resp)

	// 7. ส่ง event จบ
	writeEvent(c, "end", "ตอบคำถามทักทาย/กล่าวลาเสร็จสิ้น")
	return fullText
}

// AGENT: Generic question agent
func StreamGenericQuestionAgent(c *gin.Context, question string) {
	// 1. สร้าง system prompt
	systemPrompt := `คุณเป็นผู้ช่วยตอบคำถามทั่วไปใน BSO Space
โปรดตอบคำถามต่อไปนี้อย่างสุภาพและเป็นประโยชน์
คำถาม: ` + question

	// 2. เตรียม payload
	payload := map[string]interface{}{
		"model":  os.Getenv("AI_MODEL"),
		"stream": true,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": question},
		},
	}

	// 3. เตรียม HTTP request
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("AI_API_KEY"))
	req.Header.Set("HTTP-Referer", "https://blog.bsospace.com")
	req.Header.Set("X-Title", "https://blog.bsospace.com")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		writeErrorEvent(c, "เรียกใช้ LLM ไม่สำเร็จ")
		return
	}
	defer resp.Body.Close()

	// 4. ตั้ง header สำหรับ SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Writer.Flush()

	// 5. ส่ง event เริ่ม
	writeEvent(c, "start", "เริ่มการตอบคำถามทั่วไป")

	// 6. อ่านและส่ง stream
	streamLLMResponse(c, resp)

	// 7. ส่ง event จบ
	writeEvent(c, "end", "ตอบคำถามทั่วไปเสร็จสิ้น")
}

// parseIntentFromLLM extracts the intent keyword from LLM response.
func parseIntentFromLLM(resp string) Intent {
	resp = strings.TrimSpace(resp)

	// Try to parse as JSON (OpenRouter format)
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(resp), &parsed); err == nil && len(parsed.Choices) > 0 {
		intent := strings.TrimSpace(parsed.Choices[0].Message.Content)
		switch intent {
		case "blog_question":
			return IntentBlogQuestion
		case "summarize_post":
			return IntentSummarizePost
		case "search_blog":
			return IntentSearchBlog
		case "generic_question":
			return IntentGeneric
		case "greeting_farewell":
			return IntentGreetingFarewell
		default:
			return IntentUnknown
		}
	}

	// fallback: direct string match
	// fallback: direct string match
	resp = strings.ToLower(resp)
	resp = strings.Trim(resp, "\"") // <-- เพิ่มบรรทัดนี้เพื่อกัน "blog_question"
	switch resp {
	case "blog_question":
		return IntentBlogQuestion
	case "summarize_post":
		return IntentSummarizePost
	case "search_blog":
		return IntentSearchBlog
	case "generic_question":
		return IntentGeneric
	case "greeting_farewell":
		return IntentGreetingFarewell
	default:
		return IntentUnknown
	}

}

// SendLLMRequestToOpenRouter is a utility for agents to call OpenRouter or Ollama.
// Returns the response body as a string (for non-streaming use).
func SendLLMRequestToOpenRouter(payload map[string]interface{}, config map[string]string) (string, error) {
	body, _ := json.Marshal(payload)

	useSelfHost := config["use_self_host"] == "true"
	if useSelfHost {
		host := config["host"]
		resp, err := http.Post(host+"/api/chat", "application/json", bytes.NewBuffer(body))
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("Ollama error: %s", string(respBody))
		}
		return string(respBody), nil
	}

	apiKey := config["api_key"]
	if apiKey == "" {
		return "", fmt.Errorf("OpenRouter API key missing")
	}
	request, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+apiKey)
	request.Header.Set("HTTP-Referer", "https://blog.bsospace.com")
	request.Header.Set("X-Title", "https://blog.bsospace.com")
	client := &http.Client{}
	log.Printf("Sending intent classification to OpenRouter: %+v", payload)
	resp, err := client.Do(request)
	log.Printf("OpenRouter response status: %v", resp.StatusCode)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenRouter error: %s", string(respBody))
	}
	return string(respBody), nil
}

// BlogAgent handles blog content questions and generic questions.
type BlogAgent struct{}

// streamLLMResponse streams LLM response and calls onChunk for each content chunk.
func streamLLMResponse(c *gin.Context, resp *http.Response) string {
	reader := bufio.NewReader(resp.Body)
	var fullText strings.Builder

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				// a.logger.Info("LLM stream finished")
				writeEvent(c, "end", "done")
			} else {
				// a.logger.Error("Error reading LLM stream", zap.Error(err))
				writeErrorEvent(c, "Stream reading error")
			}
			break
		}

		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		raw := bytes.TrimSpace(line[6:])
		if len(raw) == 0 {
			continue
		}

		if bytes.Equal(raw, []byte("[DONE]")) {
			// a.writeEvent(c, "end", "done")
			break
		}

		content := parseStreamChunk(raw)
		if content != "" {
			fullText.WriteString(content)
			jsonEncoded, _ := json.Marshal(map[string]string{"text": content})
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonEncoded)
			c.Writer.Flush()
		}
	}

	return fullText.String()
}

// SummarizerAgent handles summarization requests.
type SummarizerAgent struct{}

func (a *SummarizerAgent) Handle(question string, ctx *AgentContext) (string, error) {
	if ctx.Post == nil {
		return "ไม่พบโพสต์นี้ในระบบ", nil
	}
	contextText := ctx.Post.Content
	systemPrompt := `คุณเป็นผู้ช่วยสรุปเนื้อหาบทความใน BSO Space Blog
โปรดสรุปเนื้อหาด้านล่างนี้ให้กระชับ เข้าใจง่าย และเป็นมิตร
-----
` + contextText + `
-----`

	payload := map[string]interface{}{
		"model":  os.Getenv("AI_MODEL"),
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": question},
		},
	}
	config := map[string]string{
		"api_key":       os.Getenv("AI_API_KEY"),
		"model":         os.Getenv("AI_MODEL"),
		"host":          os.Getenv("AI_HOST"),
		"use_self_host": os.Getenv("AI_SELF_HOST"),
	}
	resp, err := SendLLMRequestToOpenRouter(payload, config)
	if err != nil {
		return "เกิดข้อผิดพลาดในการเชื่อมต่อ AI", err
	}
	return strings.TrimSpace(resp), nil
}

// SearchAgent handles blog search requests.
type SearchAgent struct{}

func (a *SearchAgent) Handle(question string, ctx *AgentContext) (string, error) {
	systemPrompt := `คุณเป็นผู้ช่วยค้นหาบทความใน BSO Space Blog
โปรดค้นหาบทความที่เกี่ยวข้องกับคำถามจากฐานข้อมูลหรือ external source (mock)
ถ้าพบให้แสดงรายการชื่อบทความ ถ้าไม่พบให้ตอบว่า "ไม่พบผลลัพธ์ที่เกี่ยวข้อง"
`
	// ในอนาคตสามารถเชื่อมต่อ vector search/external API ได้
	payload := map[string]interface{}{
		"model":  os.Getenv("AI_MODEL"),
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": question},
		},
	}
	config := map[string]string{
		"api_key":       os.Getenv("AI_API_KEY"),
		"model":         os.Getenv("AI_MODEL"),
		"host":          os.Getenv("AI_HOST"),
		"use_self_host": os.Getenv("AI_SELF_HOST"),
	}
	resp, err := SendLLMRequestToOpenRouter(payload, config)
	if err != nil {
		return "เกิดข้อผิดพลาดในการเชื่อมต่อ AI", err
	}
	return strings.TrimSpace(resp), nil
}

// FallbackAgent handles unknown or unsupported intents.
type FallbackAgent struct{}

func (a *FallbackAgent) Handle(question string, ctx *AgentContext) (string, error) {
	return "ขออภัย ไม่เข้าใจคำถาม ลองใหม่อีกครั้งหรือถามในรูปแบบอื่นได้ค่ะ 😊", nil
}
