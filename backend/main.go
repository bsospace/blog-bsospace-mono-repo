package main

import (
	"rag-searchbot-backend/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/upload", handlers.UploadHandler)
	r.POST("/ask", handlers.AskHandler)

	r.Run(":8080")
}
