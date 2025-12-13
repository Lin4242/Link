package service

import (
	"context"
	"time"

	"link/internal/domain"
)

type MessageService struct {
	msgRepo  domain.MessageRepository
	convRepo domain.ConversationRepository
}

func NewMessageService(msgRepo domain.MessageRepository, convRepo domain.ConversationRepository) *MessageService {
	return &MessageService{msgRepo: msgRepo, convRepo: convRepo}
}

func (s *MessageService) Send(ctx context.Context, senderID, conversationID, encryptedContent string) (*domain.Message, error) {
	conv, err := s.convRepo.FindByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	if conv.Participant1 != senderID && conv.Participant2 != senderID {
		return nil, domain.ErrUnauthorized("not a participant")
	}

	msg := &domain.Message{
		ConversationID:   conversationID,
		SenderID:         senderID,
		EncryptedContent: encryptedContent,
	}
	if err := s.msgRepo.Create(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *MessageService) GetMessages(ctx context.Context, userID, conversationID string, limit int, before *time.Time) ([]*domain.Message, error) {
	conv, err := s.convRepo.FindByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	if conv.Participant1 != userID && conv.Participant2 != userID {
		return nil, domain.ErrUnauthorized("not a participant")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	return s.msgRepo.FindByConversation(ctx, conversationID, limit, before)
}

func (s *MessageService) Delete(ctx context.Context, userID, messageID string) (*domain.Message, error) {
	msg, err := s.msgRepo.FindByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	// Only sender can delete
	if msg.SenderID != userID {
		return nil, domain.ErrUnauthorized("only sender can delete message")
	}

	if err := s.msgRepo.Delete(ctx, messageID); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *MessageService) MarkDelivered(ctx context.Context, messageID string) error {
	return s.msgRepo.MarkDelivered(ctx, messageID)
}

func (s *MessageService) MarkRead(ctx context.Context, messageID string) error {
	return s.msgRepo.MarkRead(ctx, messageID)
}
