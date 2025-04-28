package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UploadHandler(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid file"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "Cannot open file"})
		return
	}
	defer file.Close()

	// Prepare multipart request to Python service
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, _ := w.CreateFormFile("file", fileHeader.Filename)
	io.Copy(fw, file)
	w.Close()

	// Call Python extractor
	resp, err := http.Post("http://localhost:5001/extract-text", w.FormDataContentType(), &b)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to call extractor service"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]string
	json.Unmarshal(body, &result)

	text, ok := result["text"]
	if !ok {
		c.JSON(500, gin.H{"error": "Failed to extract text"})
		return
	}

	c.JSON(200, gin.H{
		"message": "File processed successfully",
		"text":    text,
	})
}
