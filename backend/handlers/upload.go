package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"rag-searchbot-backend/internal/ollama"
	"rag-searchbot-backend/internal/storage"

	"github.com/gin-gonic/gin"
)

var isSelfHost = os.Getenv("AI_SELF_HOST") == "true"
var externalAPIKey = os.Getenv("AI_API_KEY")
var externalModel = os.Getenv("AI_MODEL")

func UploadHandler(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot open file"})
		return
	}
	defer file.Close()

	// ส่งไป extractor
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create form"})
		return
	}
	if _, err := io.Copy(part, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy file"})
		return
	}
	writer.Close()

	resp, err := http.Post("http://bobby.posyayee.com:5002/extract-text", writer.FormDataContentType(), &buffer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call extractor"})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode extractor response"})
		return
	}

	text := result.Text
	fmt.Println("Extracted text:", text)

	chunks := splitText(text, 500)
	var chunkObjs []storage.Chunk

	for _, ch := range chunks {
		embedding, err := getEmbedding(ch)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create embedding",
				"message": err.Error(),
			})
			return
		}
		float32Embedding := make([]float32, len(embedding))
		for i, v := range embedding {
			float32Embedding[i] = float32(v)
		}
		chunkObjs = append(chunkObjs, storage.Chunk{
			Text:      ch,
			Embedding: float32Embedding,
		})
	}

	storage.SaveChunksWithEmbeddings(chunkObjs)

	c.JSON(http.StatusOK, gin.H{
		"message": "File processed successfully",
		"chunks":  len(chunkObjs),
	})
}

func splitText(text string, chunkSize int) []string {
	var chunks []string
	runes := []rune(text)
	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	return chunks
}

func getEmbedding(text string) ([]float64, error) {
	// ใช้ embedding จาก Ollama local เท่านั้น
	embedding32, err := ollama.GetEmbedding(text)
	if err != nil {
		return nil, err
	}

	// แปลง []float32 → []float64 หากระบบ downstream ต้องการ
	embedding64 := make([]float64, len(embedding32))
	for i, v := range embedding32 {
		embedding64[i] = float64(v)
	}

	return embedding64, nil
}
