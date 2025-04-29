package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"rag-searchbot-backend/internal/ollama"
	"rag-searchbot-backend/internal/storage"

	"github.com/gin-gonic/gin"
)

func UploadHandler(c *gin.Context) {
	// รับไฟล์จากฟอร์ม
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

	// เตรียม multipart สำหรับส่งไป Extractor
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

	// เรียก Python Extractor API
	resp, err := http.Post("http://192.168.1.105:5002/extract-text", writer.FormDataContentType(), &buffer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call extractor"})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Text string `json:"text"`
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode extractor response"})
		return
	}

	text := result.Text

	fmt.Println("Extracted text:", text)

	// ✨ Text Chunking (500 words ต่อ chunk)
	chunks := splitText(text, 500)

	// ✨ ทำ Embedding ให้แต่ละ Chunk และเก็บเข้า Memory
	var chunkObjs []storage.Chunk
	for _, ch := range chunks {
		embedding, err := ollama.GetEmbedding(ch)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create embedding", "message": err.Error()})
			return
		}
		chunkObjs = append(chunkObjs, storage.Chunk{
			Text:      ch,
			Embedding: embedding,
		})
	}

	// Save ลง Memory
	storage.SaveChunksWithEmbeddings(chunkObjs)

	// ตอบกลับ Frontend
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
