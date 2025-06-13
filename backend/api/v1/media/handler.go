package media

import (
	"net/http"
	"rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MediaHandler struct {
	MediaService *media.MediaService
}

func NewMediaHandler(mediaService *media.MediaService) *MediaHandler {
	return &MediaHandler{
		MediaService: mediaService,
	}
}

// UploadImageHandler handles the image upload request
func (h *MediaHandler) UploadImageHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "Missing or invalid file"})
		return
	}

	postIDStr := c.PostForm("post_id")
	var postID *uuid.UUID
	if postIDStr != "" {
		uid, err := uuid.Parse(postIDStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid post_id UUID"})
			return
		}
		postID = &uid
	}

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

	// Upload
	image, err := h.MediaService.CreateMedia(file, userData, postID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": "Failed to upload image", "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "Image uploaded successfully",
		"data":    image,
	})
}
