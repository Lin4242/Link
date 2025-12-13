package handler

import (
	"time"

	"link/internal/service"

	"github.com/gofiber/fiber/v2"
)

// Notifier interface for sending WebSocket notifications
type Notifier interface {
	SendTyped(userID string, msgType string, payload interface{}) bool
}

type ConversationHandler struct {
	convSvc  *service.ConversationService
	msgSvc   *service.MessageService
	notifier Notifier
}

func NewConversationHandler(convSvc *service.ConversationService, msgSvc *service.MessageService, notifier Notifier) *ConversationHandler {
	return &ConversationHandler{convSvc: convSvc, msgSvc: msgSvc, notifier: notifier}
}

func (h *ConversationHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	conversations, err := h.convSvc.GetUserConversations(c.Context(), userID)
	if err != nil {
		return Error(c, err)
	}
	return OK(c, conversations)
}

func (h *ConversationHandler) Messages(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	convID := c.Params("id")

	limit := c.QueryInt("limit", 50)
	beforeStr := c.Query("before")

	var before *time.Time
	if beforeStr != "" {
		t, err := time.Parse(time.RFC3339, beforeStr)
		if err == nil {
			before = &t
		}
	}

	messages, err := h.msgSvc.GetMessages(c.Context(), userID, convID, limit, before)
	if err != nil {
		return Error(c, err)
	}
	return OK(c, messages)
}

func (h *ConversationHandler) DeleteMessage(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	messageID := c.Params("messageId")

	msg, err := h.msgSvc.Delete(c.Context(), userID, messageID)
	if err != nil {
		return Error(c, err)
	}

	// Notify the peer about deletion
	conv, err := h.convSvc.GetByID(c.Context(), msg.ConversationID)
	if err == nil && h.notifier != nil {
		var peerID string
		if conv.Participant1 == userID {
			peerID = conv.Participant2
		} else {
			peerID = conv.Participant1
		}
		h.notifier.SendTyped(peerID, "deleted", map[string]string{
			"id":              msg.ID,
			"conversation_id": msg.ConversationID,
		})
	}

	return OK(c, map[string]string{
		"id":              msg.ID,
		"conversation_id": msg.ConversationID,
	})
}
