package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"rag-searchbot-backend/internal/ws"
	"rag-searchbot-backend/pkg/ginctx"
	"rag-searchbot-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WebSocketHandler struct {
	manager *ws.Manager
}

func NewWebSocketHandler(manager *ws.Manager) *WebSocketHandler {
	return &WebSocketHandler{manager: manager}
}

func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	user, ok := ginctx.GetUserFromContext(c)
	if !ok || user == nil {
		response.JSONError(c, 401, "Unauthorized", "User not found in context")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.String(http.StatusInternalServerError, "Upgrade failed")
		return
	}

	h.manager.AddClient(user.ID.String(), conn)

	log.Printf("[WS] Client connected: %s (%s)", user.Email, user.ID)
	h.manager.LogAllClients()

	defer func() {
		h.manager.RemoveClient(user.ID.String())
		conn.Close()
	}()

	// Send sayhi message
	sayhiMessage := map[string]interface{}{
		"event":   "sayhi",
		"message": `Welcome to the WebSocket server! ` + user.UserName,
	}
	if messageBytes, err := json.Marshal(sayhiMessage); err == nil {
		conn.WriteMessage(websocket.TextMessage, messageBytes)
	}

	// Read loop + dump client message
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[WS] Client disconnected or error occurred: %v", err)
			break
		}

		log.Printf("[WS] Received message from user %s (type %d): %s", user.Email, messageType, string(message))
	}
}
