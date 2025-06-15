package ws

import (
	"log"
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
	client, ok := m.clients[userID]
	m.mu.RUnlock()

	if !ok || client == nil || client.Conn == nil {
		log.Printf("[WS] User %s is not connected - cannot send %s", userID, event)
		return nil
	}

	msg := map[string]interface{}{
		"event":   event,
		"payload": payload,
	}

	log.Printf("[WS] Sending to user %s - event: %s - payload: %+v", userID, event, payload)

	if err := client.Conn.WriteJSON(msg); err != nil {
		log.Printf("[WS] Failed to send to user %s: %v", userID, err)
		return err
	}

	return nil
}

func (m *Manager) GetClient(userID string) *Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, ok := m.clients[userID]
	if !ok {
		return nil
	}
	return client
}

func (m *Manager) LogAllClients() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	log.Println("[WS] Currently connected clients:")
	for userID, client := range m.clients {
		if client != nil && client.Conn != nil {
			log.Printf(" - userID: %s", userID)
		}
	}
}
