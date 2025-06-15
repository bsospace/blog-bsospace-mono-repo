package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserID string
	Conn   *websocket.Conn
}

type Manager struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*Client),
	}
}

func (m *Manager) AddClient(userID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[userID] = &Client{UserID: userID, Conn: conn}
}

func (m *Manager) RemoveClient(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, userID)
}

func (m *Manager) SendToUser(userID, event string, payload interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, ok := m.clients[userID]
	if !ok {
		return nil // user not connected
	}

	return client.Conn.WriteJSON(map[string]interface{}{
		"event":   event,
		"payload": payload,
	})
}
