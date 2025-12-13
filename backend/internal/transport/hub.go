package transport

import (
	"log/slog"
	"sync"
)

type Client interface {
	GetUserID() string
	SendStream(msg *Message) bool
	SendDatagram(msg *Message) bool
	Close()
}

type Hub struct {
	clients    map[string]Client
	mu         sync.RWMutex
	register   chan Client
	unregister chan Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]Client),
		register:   make(chan Client, 256),
		unregister: make(chan Client, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if old, ok := h.clients[c.GetUserID()]; ok {
				old.Close()
			}
			h.clients[c.GetUserID()] = c
			h.mu.Unlock()
			slog.Info("client connected", "user_id", c.GetUserID())

		case c := <-h.unregister:
			h.mu.Lock()
			if curr, ok := h.clients[c.GetUserID()]; ok && curr == c {
				delete(h.clients, c.GetUserID())
			}
			h.mu.Unlock()
			slog.Info("client disconnected", "user_id", c.GetUserID())
		}
	}
}

func (h *Hub) Send(userID string, msg *Message) bool {
	h.mu.RLock()
	c, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return false
	}
	return c.SendStream(msg)
}

// SendTyped implements the Notifier interface for HTTP handlers
func (h *Hub) SendTyped(userID string, msgType string, payload interface{}) bool {
	return h.Send(userID, &Message{Type: msgType, Payload: payload})
}

func (h *Hub) SendDatagram(userID string, msg *Message) bool {
	h.mu.RLock()
	c, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return false
	}
	return c.SendDatagram(msg)
}

func (h *Hub) IsOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

func (h *Hub) Register(c Client)   { h.register <- c }
func (h *Hub) Unregister(c Client) { h.unregister <- c }
