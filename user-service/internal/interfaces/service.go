package interfaces

import (
	"context"
	"user_service/internal/application/dto"
)

type UserService interface {
	Create(ctx context.Context, user *dto.CreateUser) (*dto.User, error)
	GetById(ctx context.Context, id string) (*dto.User, error)
	Paginate(ctx context.Context, page, limit int) ([]dto.User, error)

	Follow(ctx context.Context, id, followerID string) error
	Unfollow(ctx context.Context, id, followerID string) error
	Followers(ctx context.Context, id string, page, limit int) ([]dto.Follower, error)
	Following(ctx context.Context, id string, page, limit int) ([]dto.Follower, error)
}
