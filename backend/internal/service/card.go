package service

import (
	"context"

	"link/internal/domain"
	"link/internal/pkg/cardtoken"
)

type CardService struct {
	cardRepo    domain.CardRepository
	sessionRepo domain.SessionRepository
	tokenGen    *cardtoken.Generator
}

func NewCardService(cardRepo domain.CardRepository, sessionRepo domain.SessionRepository, tokenGen *cardtoken.Generator) *CardService {
	return &CardService{cardRepo: cardRepo, sessionRepo: sessionRepo, tokenGen: tokenGen}
}

func (s *CardService) CheckCard(ctx context.Context, token string) (*domain.CardCheckResult, error) {
	// First check if already registered
	card, err := s.cardRepo.FindByToken(ctx, token)
	if err == nil && card != nil {
		if card.Status == domain.CardStatusRevoked {
			return &domain.CardCheckResult{Status: "revoked"}, nil
		}
		var warning string
		if card.CardType == domain.CardTypeBackup {
			warning = "此為備援卡，使用後主卡將失效"
		}
		return &domain.CardCheckResult{
			Status:   string(card.CardType),
			UserID:   &card.UserID,
			CardType: &card.CardType,
			Warning:  &warning,
		}, nil
	}

	// Validate token format using HMAC
	_, _, err = s.tokenGen.ParseToken(token)
	if err != nil {
		return &domain.CardCheckResult{Status: "invalid_token"}, nil
	}

	// Check if paired card is already registered
	pairedToken, _ := s.tokenGen.GetPairedToken(token)
	pairedCard, _ := s.cardRepo.FindByToken(ctx, pairedToken)

	if pairedCard != nil {
		// Paired card is already registered - this pair can't be used
		return &domain.CardCheckResult{Status: "pair_already_registered"}, nil
	}

	// For neutral format tokens, just return can_register
	// The frontend will handle the primary/backup assignment based on scan order
	return &domain.CardCheckResult{Status: "can_register", PairedToken: &pairedToken}, nil
}

// ValidateTokenPair checks if the primary and backup tokens are a valid pair
func (s *CardService) ValidateTokenPair(primaryToken, backupToken string) error {
	// Check if they form a valid pair
	if !s.tokenGen.ArePaired(primaryToken, backupToken) {
		return domain.ErrValidation("這兩張卡不是配對的卡片")
	}

	// For legacy format (P/B), check primary is actually primary
	// For neutral format (1/2), no check needed - roles are assigned by scan order
	if !s.tokenGen.IsNeutralFormat(primaryToken) {
		if !s.tokenGen.IsPrimary(primaryToken) {
			return domain.ErrValidation("第一張卡必須是主卡")
		}
	}

	return nil
}

// GetPairedToken returns the paired token for a given token
func (s *CardService) GetPairedToken(token string) (string, error) {
	return s.tokenGen.GetPairedToken(token)
}

// IsPrimaryToken checks if the token is a primary card
func (s *CardService) IsPrimaryToken(token string) bool {
	return s.tokenGen.IsPrimary(token)
}

func (s *CardService) RevokeWithBackupCard(ctx context.Context, backupCardID, userID string) error {
	primaryCard, err := s.cardRepo.FindActiveByUserAndType(ctx, userID, domain.CardTypePrimary)
	if err == nil && primaryCard != nil {
		if err := s.cardRepo.Revoke(ctx, primaryCard.ID); err != nil {
			return err
		}
	}

	if err := s.cardRepo.PromoteBackupToPrimary(ctx, backupCardID); err != nil {
		return err
	}

	return s.sessionRepo.RevokeAllByUser(ctx, userID)
}

func (s *CardService) GetUserCards(ctx context.Context, userID string) ([]*domain.Card, error) {
	return s.cardRepo.FindByUserID(ctx, userID)
}
