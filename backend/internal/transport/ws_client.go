package transport

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

type WSClient struct {
	userID  string
	conn    *websocket.Conn
	hub     *Hub
	handler *Handler
	send    chan []byte
	mu      sync.Mutex
}

func NewWSClient(userID string, conn *websocket.Conn, hub *Hub, handler *Handler) *WSClient {
	return &WSClient{
		userID:  userID,
		conn:    conn,
		hub:     hub,
		handler: handler,
		send:    make(chan []byte, 256),
	}
}

func (c *WSClient) GetUserID() string { return c.userID }

func (c *WSClient) SendStream(msg *Message) bool {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal message", "err", err)
		return false
	}
	slog.Info("SendStream: putting message in channel", "type", msg.Type, "user_id", c.userID, "data_len", len(data))
	select {
	case c.send <- data:
		slog.Info("SendStream: message queued successfully")
		return true
	default:
		slog.Warn("SendStream: channel full, dropping message")
		return false
	}
}

func (c *WSClient) SendDatagram(msg *Message) bool {
	return c.SendStream(msg)
}

func (c *WSClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.Close()
}

func (c *WSClient) Run(ctx context.Context) {
	go c.writePump()
	c.readPump(ctx)
	c.hub.Unregister(c)
}

func (c *WSClient) readPump(ctx context.Context) {
	defer c.conn.Close()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg struct {
			Type    string          `json:"t"`
			Payload json.RawMessage `json:"p"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			slog.Error("failed to unmarshal ws message", "err", err, "data", string(data))
			continue
		}

		slog.Info("WebSocket message received", "type", msg.Type, "user_id", c.userID)

		switch msg.Type {
		case TypeMessage:
			c.handler.HandleMessage(ctx, c.userID, msg.Payload)
		case TypeRead:
			c.handler.HandleRead(ctx, c.userID, msg.Payload)
		case TypeTyping:
			var p struct {
				To             string `json:"to"`
				ConversationID string `json:"conversation_id"`
			}
			if json.Unmarshal(msg.Payload, &p) == nil {
				c.hub.Send(p.To, &Message{
					Type:    TypeTyping,
					Payload: map[string]string{"from": c.userID, "conversation_id": p.ConversationID},
				})
			}
		}
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			slog.Info("writePump: writing message to websocket", "user_id", c.userID, "msg", string(msg))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				slog.Error("ws write error", "err", err)
				return
			}
			slog.Info("writePump: message written successfully")
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
