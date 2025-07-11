package handler

import (
	"net/http"
	"rag-searchbot-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// MeHandler คืนค่าข้อมูลของผู้ใช้ที่เข้าสู่ระบบ (เฉพาะบาง field)
func Me(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not found in context",
		})
		return
	}

	userData, ok := user.(*models.User)
	if !ok || userData == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid user data",
		})
		return
	}

	warpKey, exists := c.Get("warp_key")
	if !exists || warpKey == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Warp key not found in context",
		})
		return
	}

	resp := MapResponse(userData)

	resp.WarpKey = warpKey.(string)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User info fetched successfully",
		"data":    resp,
	})
}
