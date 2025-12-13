package domain

import (
	"context"
	"time"
)

type CardType string
type CardStatus string

const (
	CardTypePrimary CardType = "primary"
	CardTypeBackup  CardType = "backup"

	CardStatusActive  CardStatus = "active"
	CardStatusRevoked CardStatus = "revoked"
)

type Card struct {
	ID          string
	UserID      string
	CardToken   string
	CardType    CardType
	Status      CardStatus
	CreatedAt   time.Time
	ActivatedAt *time.Time
	RevokedAt   *time.Time
}

type CardPair struct {
	ID           string
	PrimaryToken string
	BackupToken  *string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

type CardRepository interface {
	FindByToken(ctx context.Context, token string) (*Card, error)
	FindByUserID(ctx context.Context, userID string) ([]*Card, error)
	FindActiveByUserAndType(ctx context.Context, userID string, cardType CardType) (*Card, error)

	Create(ctx context.Context, card *Card) error
	Revoke(ctx context.Context, cardID string) error
	PromoteBackupToPrimary(ctx context.Context, cardID string) error

	CreatePair(ctx context.Context, primaryToken string) (*CardPair, error)
	FindPairByPrimaryToken(ctx context.Context, token string) (*CardPair, error)
	FindPairByBackupToken(ctx context.Context, token string) (*CardPair, error)
	UpdatePairBackupToken(ctx context.Context, pairID, backupToken string) error
	DeletePair(ctx context.Context, pairID string) error
	CleanupExpiredPairs(ctx context.Context) error
}

type CardCheckResult struct {
	Status      string    `json:"status"`
	UserID      *string   `json:"user_id,omitempty"`
	Nickname    *string   `json:"nickname,omitempty"`
	CardType    *CardType `json:"card_type,omitempty"`
	PairID      *string   `json:"pair_id,omitempty"`
	Warning     *string   `json:"warning,omitempty"`
	PairedToken *string   `json:"paired_token,omitempty"`
}
