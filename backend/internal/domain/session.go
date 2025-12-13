package domain

import (
	"context"
	"time"
)

type Session struct {
	ID        string
	UserID    string
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	FindByTokenHash(ctx context.Context, hash string) (*Session, error)
	RevokeAllByUser(ctx context.Context, userID string) error
	Revoke(ctx context.Context, id string) error
	CleanupExpired(ctx context.Context) error
}
