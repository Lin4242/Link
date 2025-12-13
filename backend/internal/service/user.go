package service

import (
	"context"

	"link/internal/domain"
)

type UserService struct {
	userRepo domain.UserRepository
}

func NewUserService(userRepo domain.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *UserService) GetPublicKey(ctx context.Context, id string) (string, error) {
	return s.userRepo.GetPublicKey(ctx, id)
}

func (s *UserService) Update(ctx context.Context, user *domain.User) error {
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) UpdateLastSeen(ctx context.Context, id string) error {
	return s.userRepo.UpdateLastSeen(ctx, id)
}

func (s *UserService) Search(ctx context.Context, query string, limit int) ([]*domain.User, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.userRepo.Search(ctx, query, limit)
}
