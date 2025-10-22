package ai

import (
	"fmt"

	"rag-searchbot-backend/internal/models"
	"rag-searchbot-backend/internal/post"

	"github.com/gin-gonic/gin"
)

// Intent represents the type of user question.

// Agent is the interface for all agents.
type Agent interface {
	Handle(question string, ctx *AgentContext) (string, error)
}

// AgentContext provides context for agent processing.
type AgentContext struct {
	Post    *models.Post
	User    *models.User
	PosRepo post.PostRepositoryInterface
	// เพิ่ม field อื่นๆ ได้ เช่น Embeddings, Logger, etc.
}

// ErrInsufficientContext is returned by BlogAgent when RAG context is not enough.
var ErrInsufficientContext = fmt.Errorf("insufficient context for RAG")

// Helper methods
func writeEvent(c *gin.Context, event, data string) {
	fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data)
	c.Writer.Flush()
}

func writeErrorEvent(c *gin.Context, message string) {
	fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", message)
	c.Writer.Flush()
}
