package media

import (
	"fmt"
	"net/http"
	"rag-searchbot-backend/internal/media"
	"rag-searchbot-backend/internal/models"

	"github.com/gin-gonic/gin"
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
	var form media.UploadMediaForm

	if err := c.ShouldBind(&form); err != nil {
		c.JSON(400, gin.H{"error": "Invalid form input"})
		return
	}
	user, exists := c.Get("user")
	if !exists || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not found in context",
		})
		return
	}

	fmt.Println("User in context:", user)

	userData, ok := user.(*models.User)
	if !ok || userData == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid user data",
		})
		return
	}

	// Call service to handle upload
	image, err := h.MediaService.CreateMedia(form.File, userData)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": "Failed to upload image", "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "Image uploaded successfully",
		"data":    image,
	})
}
