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

	// map data to response
	resp := struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Avatar    string `json:"avatar"`
		Role      string `json:"role"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		ID:        userData.ID.String(),
		Email:     userData.Email,
		Avatar:    userData.Avatar,
		Role:      string(userData.Role),
		CreatedAt: userData.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: userData.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User info fetched successfully",
		"data":    resp,
	})
}
