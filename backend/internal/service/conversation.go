package service

import (
	"context"

	"link/internal/domain"
)

type ConversationService struct {
	convRepo domain.ConversationRepository
}

func NewConversationService(convRepo domain.ConversationRepository) *ConversationService {
	return &ConversationService{convRepo: convRepo}
}

func (s *ConversationService) GetOrCreate(ctx context.Context, userA, userB string) (*domain.Conversation, error) {
	return s.convRepo.GetOrCreate(ctx, userA, userB)
}

func (s *ConversationService) GetByID(ctx context.Context, id string) (*domain.Conversation, error) {
	return s.convRepo.FindByID(ctx, id)
}

func (s *ConversationService) GetUserConversations(ctx context.Context, userID string) ([]*domain.ConversationWithPeer, error) {
	return s.convRepo.FindByUser(ctx, userID)
}
