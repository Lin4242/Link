package service

import (
	"context"

	"link/internal/domain"
)

type FriendshipService struct {
	friendRepo domain.FriendshipRepository
	userRepo   domain.UserRepository
}

func NewFriendshipService(friendRepo domain.FriendshipRepository, userRepo domain.UserRepository) *FriendshipService {
	return &FriendshipService{friendRepo: friendRepo, userRepo: userRepo}
}

func (s *FriendshipService) SendRequest(ctx context.Context, requesterID, addresseeID string) (*domain.Friendship, error) {
	if requesterID == addresseeID {
		return nil, domain.ErrSelfFriendRequest
	}

	if _, err := s.userRepo.FindByID(ctx, addresseeID); err != nil {
		return nil, domain.ErrUserNotFound
	}

	existing, err := s.friendRepo.FindByUsers(ctx, requesterID, addresseeID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.Status == domain.FriendshipAccepted {
			return nil, domain.ErrAlreadyFriends
		}
		return nil, domain.ErrConflict("已有待處理的好友請求")
	}

	friendship := &domain.Friendship{
		RequesterID: requesterID,
		AddresseeID: addresseeID,
		Status:      domain.FriendshipPending,
	}
	if err := s.friendRepo.Create(ctx, friendship); err != nil {
		return nil, err
	}
	return friendship, nil
}

func (s *FriendshipService) Accept(ctx context.Context, userID, friendshipID string) error {
	return s.friendRepo.UpdateStatus(ctx, friendshipID, domain.FriendshipAccepted)
}

func (s *FriendshipService) Reject(ctx context.Context, userID, friendshipID string) error {
	return s.friendRepo.Delete(ctx, friendshipID)
}

func (s *FriendshipService) Remove(ctx context.Context, userID, friendshipID string) error {
	return s.friendRepo.Delete(ctx, friendshipID)
}

func (s *FriendshipService) GetFriends(ctx context.Context, userID string) ([]*domain.FriendWithUser, error) {
	return s.friendRepo.FindFriends(ctx, userID)
}

func (s *FriendshipService) GetPendingRequests(ctx context.Context, userID string) ([]*domain.FriendWithUser, error) {
	return s.friendRepo.FindPendingRequests(ctx, userID)
}
