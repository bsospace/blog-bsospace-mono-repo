package ai

import (
	"net/http"
	"rag-searchbot-backend/internal/ai"
	"rag-searchbot-backend/pkg/ginctx"
	"rag-searchbot-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AIHandler struct {
	AIService *ai.AIService
}

func NewAIHandler(aiService *ai.AIService) *AIHandler {
	return &AIHandler{
		AIService: aiService,
	}
}

func (a *AIHandler) OpenAIMode(c *gin.Context) {

	postID := c.Param("post_id")

	if postID == "" {
		response.JSONSuccess(c, http.StatusBadRequest, "Bad request", "Post id required")
		return
	}

	user, ok := ginctx.GetUserFromContext(c)
	if !ok || user == nil {
		response.JSONError(c, http.StatusUnauthorized, "User not found in context", "User context is missing")
	}

	// Use existingPost in your further logic here
	result, err := a.AIService.OpenAIMode(postID, user)

	if err != nil && err == gorm.ErrRecordNotFound {
		response.JSONError(c, http.StatusNotFound, "Not found", "Post not found!")
		return
	}

	if err != nil {
		response.JSONError(c, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	if result {
		response.JSONSuccess(c, http.StatusOK, "Success", "AI mode in queue")
		return
	}

	response.JSONSuccess(c, http.StatusOK, "Success", "AI mode disabled")
}
