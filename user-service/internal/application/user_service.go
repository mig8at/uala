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

func (s *userService) GetById(ctx context.Context, id string) (*dto.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	getUser, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	userDTO := &dto.User{}
	if err := copier.Copy(userDTO, getUser); err != nil {
		return nil, err
	}

	return userDTO, nil
}

func (s *userService) Paginate(ctx context.Context, page, limit int) ([]dto.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	users, err := s.repo.Paginate(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	usersDTO := make([]dto.User, 0, len(users))

	for _, user := range users {
		userDTO := dto.User{}
		if err := copier.Copy(&userDTO, &user); err != nil {
			return nil, err
		}
		usersDTO = append(usersDTO, userDTO)
	}

	return usersDTO, nil
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

func (s *userService) Following(ctx context.Context, id string, page, limit int) ([]dto.Follower, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	users, err := s.repo.Following(ctx, id, page, limit)
	if err != nil {
		return nil, err
	}

	followers := make([]dto.Follower, 0, len(users))

	for _, user := range users {
		followerDto := dto.Follower{}
		if err := copier.Copy(&followerDto, &user); err != nil {
			return nil, err
		}
		followers = append(followers, followerDto)
	}

	return followers, nil
}

func (s *userService) Followers(ctx context.Context, id string, page, limit int) ([]dto.Follower, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	users, err := s.repo.Followers(ctx, id, page, limit)
	if err != nil {
		return nil, err
	}

	followers := make([]dto.Follower, 0, len(users))

	for _, user := range users {
		followerDto := dto.Follower{}
		if err := copier.Copy(&followerDto, &user); err != nil {
			return nil, err
		}
		followers = append(followers, followerDto)
	}

	return followers, nil
}
