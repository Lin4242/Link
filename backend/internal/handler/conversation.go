package handler

import (
	"time"

	"link/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ConversationHandler struct {
	convSvc *service.ConversationService
	msgSvc  *service.MessageService
}

func NewConversationHandler(convSvc *service.ConversationService, msgSvc *service.MessageService) *ConversationHandler {
	return &ConversationHandler{convSvc: convSvc, msgSvc: msgSvc}
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
