package domain

import (
	"context"
	"time"
)

type User struct {
	ID           string     `json:"id"`
	PasswordHash string     `json:"-"`
	Nickname     string     `json:"nickname"`
	PublicKey    string     `json:"public_key"`
	AvatarURL    *string    `json:"avatar_url"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastSeenAt   *time.Time `json:"last_seen_at"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	GetPublicKey(ctx context.Context, id string) (string, error)
	Update(ctx context.Context, user *User) error
	UpdateLastSeen(ctx context.Context, id string) error
	Search(ctx context.Context, query string, limit int) ([]*User, error)
}
