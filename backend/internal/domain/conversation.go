package domain

import (
	"context"
	"time"
)

type Conversation struct {
	ID            string     `json:"id"`
	Participant1  string     `json:"participant_1"`
	Participant2  string     `json:"participant_2"`
	LastMessageAt *time.Time `json:"last_message_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ConversationWithPeer struct {
	Conversation
	Peer        *User `json:"peer"`
	UnreadCount int   `json:"unread_count"`
}

type ConversationRepository interface {
	Create(ctx context.Context, c *Conversation) error
	FindByID(ctx context.Context, id string) (*Conversation, error)
	FindByParticipants(ctx context.Context, userA, userB string) (*Conversation, error)
	FindByUser(ctx context.Context, userID string) ([]*ConversationWithPeer, error)
	GetOrCreate(ctx context.Context, userA, userB string) (*Conversation, error)
}
