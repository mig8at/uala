package interfaces

import (
	"context"
	"user_service/internal/application/dto"
)

type UserService interface {
	Create(ctx context.Context, user *dto.CreateUser) (*dto.User, error)
	Follow(ctx context.Context, id, followerID string) error
	Unfollow(ctx context.Context, id, followerID string) error
}
