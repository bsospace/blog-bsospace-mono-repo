package handlers

import "github.com/gin-gonic/gin"

func AskHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Ask endpoint ready"})
}
