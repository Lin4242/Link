package transport

import (
	"context"
	"encoding/json"
	"log/slog"

	"link/internal/domain"
	"link/internal/service"
)

type Handler struct {
	hub     *Hub
	msgSvc  *service.MessageService
	convSvc *service.ConversationService
}

func NewHandler(hub *Hub, msgSvc *service.MessageService, convSvc *service.ConversationService) *Handler {
	return &Handler{hub: hub, msgSvc: msgSvc, convSvc: convSvc}
}

type SendMessagePayload struct {
	To               string `json:"to"`
	ConversationID   string `json:"conversation_id"`
	EncryptedContent string `json:"encrypted_content"`
	TempID           string `json:"temp_id"`
}

func (h *Handler) HandleMessage(ctx context.Context, senderID string, payload json.RawMessage) {
	slog.Info("HandleMessage called", "sender_id", senderID, "payload", string(payload))

	var p SendMessagePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		slog.Error("failed to unmarshal message", "err", err)
		return
	}
	slog.Info("Message parsed", "to", p.To, "temp_id", p.TempID)

	slog.Info("Calling GetOrCreate conversation")
	conv, err := h.convSvc.GetOrCreate(ctx, senderID, p.To)
	if err != nil {
		slog.Error("failed to get/create conversation", "err", err)
		return
	}
	slog.Info("Conversation retrieved", "conv_id", conv.ID)

	slog.Info("Calling msgSvc.Send")
	msg, err := h.msgSvc.Send(ctx, senderID, conv.ID, p.EncryptedContent)
	if err != nil {
		slog.Error("failed to send message", "err", err)
		return
	}
	slog.Info("Message saved", "msg_id", msg.ID)

	outMsg := &Message{
		Type: TypeMessage,
		Payload: map[string]interface{}{
			"id":                msg.ID,
			"conversation_id":   conv.ID,
			"sender_id":         senderID,
			"encrypted_content": p.EncryptedContent,
			"created_at":        msg.CreatedAt,
		},
	}

	// Always send delivery confirmation back to sender with the saved message details
	slog.Info("Sending delivery confirmation to sender", "sender_id", senderID, "temp_id", p.TempID)
	delivered := h.hub.Send(senderID, &Message{
		Type: TypeDelivered,
		Payload: map[string]interface{}{
			"temp_id": p.TempID,
			"message": map[string]interface{}{
				"id":                msg.ID,
				"conversation_id":   conv.ID,
				"sender_id":         senderID,
				"encrypted_content": p.EncryptedContent,
				"created_at":        msg.CreatedAt,
			},
		},
	})
	slog.Info("Delivery confirmation sent", "success", delivered)

	// If recipient is online, send the message and mark as delivered
	if h.hub.Send(p.To, outMsg) {
		slog.Info("Message forwarded to recipient", "to", p.To)
		_ = h.msgSvc.MarkDelivered(ctx, msg.ID)
	}
	slog.Info("HandleMessage completed")
}

type ReadPayload struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
}

func (h *Handler) HandleRead(ctx context.Context, userID string, payload json.RawMessage) {
	var p ReadPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return
	}

	if err := h.msgSvc.MarkRead(ctx, p.MessageID); err != nil {
		return
	}

	conv, err := h.convSvc.GetByID(ctx, p.ConversationID)
	if err != nil {
		return
	}

	var peerID string
	if conv.Participant1 == userID {
		peerID = conv.Participant2
	} else {
		peerID = conv.Participant1
	}

	h.hub.Send(peerID, &Message{
		Type:    TypeRead,
		Payload: map[string]string{"message_id": p.MessageID, "by": userID},
	})
}

func (h *Handler) NotifyOnline(userID string, friends []*domain.FriendWithUser) {
	for _, f := range friends {
		h.hub.Send(f.Friend.ID, &Message{
			Type:    TypeOnline,
			Payload: map[string]string{"user_id": userID},
		})
	}
}

func (h *Handler) NotifyOffline(userID string, friends []*domain.FriendWithUser) {
	for _, f := range friends {
		h.hub.Send(f.Friend.ID, &Message{
			Type:    TypeOffline,
			Payload: map[string]string{"user_id": userID},
		})
	}
}
