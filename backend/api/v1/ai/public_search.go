package ai

import (
	"encoding/json"
	"errors"
	"net/http"
	internalAI "rag-searchbot-backend/internal/ai"
	"rag-searchbot-backend/internal/models"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	publicSearchWindow        = time.Minute
	publicSearchLimit         = 10
	maxPublicSearchQueryRunes = 500
)

type publicSearchWindowState struct {
	startedAt time.Time
	count     int
}

// ponytail: process-local limiter keeps authenticated web search bounded; move
// this to Redis or the edge when the API runs across multiple instances.
type publicSearchRateLimiter struct {
	mu      sync.Mutex
	clients map[string]publicSearchWindowState
}

func newPublicSearchRateLimiter() *publicSearchRateLimiter {
	return &publicSearchRateLimiter{clients: make(map[string]publicSearchWindowState)}
}

func (l *publicSearchRateLimiter) allow(clientKey string) bool {
	now := time.Now()
	if clientKey == "" {
		clientKey = "unknown"
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for key, state := range l.clients {
		if now.Sub(state.startedAt) >= publicSearchWindow {
			delete(l.clients, key)
		}
	}

	state, ok := l.clients[clientKey]
	if !ok || now.Sub(state.startedAt) >= publicSearchWindow {
		l.clients[clientKey] = publicSearchWindowState{startedAt: now, count: 1}
		return true
	}
	if state.count >= publicSearchLimit {
		return false
	}

	state.count++
	l.clients[clientKey] = state
	return true
}

type publicWebSearchRequest struct {
	Query string `json:"query"`
}

func (a *AIHandler) WebSearch(c *gin.Context) {
	clientKey := c.ClientIP()
	if authenticatedUser, ok := c.Get("user"); ok {
		if user, ok := authenticatedUser.(*models.User); ok && user != nil {
			clientKey = user.ID.String()
		}
	}
	if a.publicSearchLimiter == nil || !a.publicSearchLimiter.allow(clientKey) {
		c.Header("Retry-After", "60")
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "web search rate limit exceeded"})
		return
	}

	var req publicWebSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid search request"})
		return
	}
	query := strings.TrimSpace(req.Query)
	if query == "" || utf8.RuneCountInString(query) > maxPublicSearchQueryRunes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query must be between 1 and 500 characters"})
		return
	}

	post, err := a.PosRepo.GetByID(c.Param("post_id"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not available for AI search"})
		return
	}
	if err != nil {
		a.logger.Error("Failed to load post for public web search", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load post"})
		return
	}
	if post == nil || !post.Published || post.PublishedAt == nil || post.Status != models.PostPublished || !post.AIChatOpen {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not available for AI search"})
		return
	}

	result, err := a.agentAgentToolWebSearchService.SearchExternalWeb(query)
	if err != nil {
		a.logger.Error("Public web search failed", zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "web search unavailable"})
		return
	}

	var payload internalAI.ChatSearchPayload
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		a.logger.Error("Public web search returned invalid payload", zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "web search returned invalid data"})
		return
	}

	c.JSON(http.StatusOK, payload)
}
