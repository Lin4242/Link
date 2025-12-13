package domain

import (
	"context"
	"time"
)

type Message struct {
	ID               string     `json:"id"`
	ConversationID   string     `json:"conversation_id"`
	SenderID         string     `json:"sender_id"`
	EncryptedContent string     `json:"encrypted_content"`
	CreatedAt        time.Time  `json:"created_at"`
	DeliveredAt      *time.Time `json:"delivered_at"`
	ReadAt           *time.Time `json:"read_at"`
}

type MessageRepository interface {
	Create(ctx context.Context, msg *Message) error
	FindByConversation(ctx context.Context, convID string, limit int, before *time.Time) ([]*Message, error)
	MarkDelivered(ctx context.Context, id string) error
	MarkRead(ctx context.Context, id string) error
}
