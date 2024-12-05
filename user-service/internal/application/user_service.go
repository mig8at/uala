package application

import (
	"context"
	"time"
	"user_service/internal/application/dto"
	"user_service/internal/interfaces"

	"github.com/jinzhu/copier"
)

type userService struct {
	repo interfaces.UserRepository
}

func NewService(repo interfaces.UserRepository) interfaces.UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) Create(ctx context.Context, user *dto.CreateUser) (*dto.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	newUser, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	userDTO := &dto.User{}
	if err := copier.Copy(userDTO, newUser); err != nil {
		return nil, err
	}

	return userDTO, nil
}

func (s *userService) Follow(ctx context.Context, id, followerID string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return s.repo.Follow(ctx, id, followerID)
}

func (s *userService) Unfollow(ctx context.Context, id, followerID string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return s.repo.Unfollow(ctx, id, followerID)
}
