package handlers

import "github.com/gin-gonic/gin"

func UploadHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Upload endpoint ready"})
}
