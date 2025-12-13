package domain

import (
	"context"
	"time"
)

type FriendshipStatus string

const (
	FriendshipPending  FriendshipStatus = "pending"
	FriendshipAccepted FriendshipStatus = "accepted"
)

type Friendship struct {
	ID          string           `json:"id"`
	RequesterID string           `json:"requester_id"`
	AddresseeID string           `json:"addressee_id"`
	Status      FriendshipStatus `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type FriendWithUser struct {
	Friendship
	Friend *User `json:"friend"`
}

type FriendshipRepository interface {
	Create(ctx context.Context, f *Friendship) error
	FindByUsers(ctx context.Context, userA, userB string) (*Friendship, error)
	FindFriends(ctx context.Context, userID string) ([]*FriendWithUser, error)
	FindPendingRequests(ctx context.Context, userID string) ([]*FriendWithUser, error)
	UpdateStatus(ctx context.Context, id string, status FriendshipStatus) error
	Delete(ctx context.Context, id string) error
}
