package service

import (
	"context"

	"link/internal/domain"
	"link/internal/pkg/cardtoken"
	"link/internal/pkg/password"
	"link/internal/pkg/token"
)

type AuthService struct {
	userRepo    domain.UserRepository
	cardRepo    domain.CardRepository
	sessionRepo domain.SessionRepository
	tokenMgr    *token.Manager
	cardTokenGen *cardtoken.Generator
}

type RegisterInput struct {
	PrimaryToken string
	BackupToken  string
	Password     string
	Nickname     string
	PublicKey    string
}

type AuthResponse struct {
	User  *domain.User `json:"user"`
	Token string       `json:"token"`
}

func NewAuthService(
	userRepo domain.UserRepository,
	cardRepo domain.CardRepository,
	sessionRepo domain.SessionRepository,
	tokenMgr *token.Manager,
	cardTokenGen *cardtoken.Generator,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		cardRepo:     cardRepo,
		sessionRepo:  sessionRepo,
		tokenMgr:     tokenMgr,
		cardTokenGen: cardTokenGen,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthResponse, error) {
	// Validate primary token format
	_, _, err := s.cardTokenGen.ParseToken(input.PrimaryToken)
	if err != nil {
		return nil, domain.ErrValidation("無效的卡片")
	}

	// For legacy format (P/B), ensure primary token is actually primary
	// For neutral format (1/2), any card can be primary (determined by scan order)
	if !s.cardTokenGen.IsNeutralFormat(input.PrimaryToken) {
		if !s.cardTokenGen.IsPrimary(input.PrimaryToken) {
			return nil, domain.ErrValidation("第一張卡必須是主卡")
		}
	}

	// If backup token provided, validate pairing
	hasBackup := input.BackupToken != ""
	if hasBackup {
		if !s.cardTokenGen.ArePaired(input.PrimaryToken, input.BackupToken) {
			return nil, domain.ErrValidation("主卡和副卡不是配對的卡片")
		}
		existingBackup, _ := s.cardRepo.FindByToken(ctx, input.BackupToken)
		if existingBackup != nil {
			return nil, domain.ErrConflict("副卡已被註冊")
		}
	}

	// Check primary card is not already registered
	existingPrimary, _ := s.cardRepo.FindByToken(ctx, input.PrimaryToken)
	if existingPrimary != nil {
		return nil, domain.ErrConflict("主卡已被註冊")
	}

	hash, err := password.Hash(input.Password)
	if err != nil {
		return nil, domain.ErrInternal()
	}

	user := &domain.User{
		PasswordHash: hash,
		Nickname:     input.Nickname,
		PublicKey:    input.PublicKey,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	primaryCard := &domain.Card{
		UserID:    user.ID,
		CardToken: input.PrimaryToken,
		CardType:  domain.CardTypePrimary,
		Status:    domain.CardStatusActive,
	}
	if err := s.cardRepo.Create(ctx, primaryCard); err != nil {
		return nil, err
	}

	// Only create backup card if token provided
	if hasBackup {
		backupCard := &domain.Card{
			UserID:    user.ID,
			CardToken: input.BackupToken,
			CardType:  domain.CardTypeBackup,
			Status:    domain.CardStatusActive,
		}
		if err := s.cardRepo.Create(ctx, backupCard); err != nil {
			return nil, err
		}
	}

	tokenStr, err := s.tokenMgr.Generate(user.ID)
	if err != nil {
		return nil, domain.ErrInternal()
	}

	return &AuthResponse{User: user, Token: tokenStr}, nil
}

func (s *AuthService) Login(ctx context.Context, cardToken, pwd string) (*AuthResponse, error) {
	card, err := s.cardRepo.FindByToken(ctx, cardToken)
	if err != nil || card == nil {
		return nil, domain.ErrUserNotFound
	}

	if card.Status == domain.CardStatusRevoked {
		return nil, domain.ErrUnauthorized("此卡片已失效")
	}

	if card.CardType == domain.CardTypeBackup {
		return nil, domain.ErrValidation("請使用主卡登入，或使用附卡撤銷流程")
	}

	user, err := s.userRepo.FindByID(ctx, card.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	ok, err := password.Verify(pwd, user.PasswordHash)
	if err != nil || !ok {
		return nil, domain.ErrInvalidPassword
	}

	tokenStr, _ := s.tokenMgr.Generate(user.ID)
	return &AuthResponse{User: user, Token: tokenStr}, nil
}

func (s *AuthService) LoginWithBackupCard(ctx context.Context, cardToken, pwd string) (*AuthResponse, error) {
	card, err := s.cardRepo.FindByToken(ctx, cardToken)
	if err != nil || card == nil {
		return nil, domain.ErrUserNotFound
	}

	if card.Status == domain.CardStatusRevoked {
		return nil, domain.ErrUnauthorized("此卡片已失效")
	}

	if card.CardType != domain.CardTypeBackup {
		return nil, domain.ErrValidation("此為主卡，請使用一般登入")
	}

	user, err := s.userRepo.FindByID(ctx, card.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	ok, err := password.Verify(pwd, user.PasswordHash)
	if err != nil || !ok {
		return nil, domain.ErrInvalidPassword
	}

	cardSvc := &CardService{cardRepo: s.cardRepo, sessionRepo: s.sessionRepo}
	if err := cardSvc.RevokeWithBackupCard(ctx, card.ID, user.ID); err != nil {
		return nil, err
	}

	tokenStr, _ := s.tokenMgr.Generate(user.ID)
	return &AuthResponse{User: user, Token: tokenStr}, nil
}
