package transport

import (
	"context"
	"log/slog"
	"strings"

	"link/internal/pkg/token"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type Server struct {
	hub      *Hub
	handler  *Handler
	tokenMgr *token.Manager
}

func NewServer(hub *Hub, handler *Handler, tokenMgr *token.Manager) *Server {
	return &Server{hub: hub, handler: handler, tokenMgr: tokenMgr}
}

func (s *Server) SetupRoutes(app *fiber.App) {
	app.Use("/ws", func(c *fiber.Ctx) error {
		slog.Info("WebSocket upgrade request received", "ip", c.IP())
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		slog.Warn("Not a WebSocket upgrade request")
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		slog.Info("WebSocket connection attempt")
		auth := c.Query("token")
		if auth == "" {
			auth = c.Headers("Authorization")
			auth = strings.TrimPrefix(auth, "Bearer ")
		}

		claims, err := s.tokenMgr.Verify(auth)
		if err != nil {
			slog.Warn("WebSocket auth failed", "error", err)
			c.Close()
			return
		}

		slog.Info("WebSocket authenticated", "user_id", claims.UserID)
		client := NewWSClient(claims.UserID, c, s.hub, s.handler)
		s.hub.Register(client)
		client.Run(context.Background())
	}))
}
