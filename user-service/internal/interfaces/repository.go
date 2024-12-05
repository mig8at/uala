package interfaces

import (
	"context"
	"user_service/internal/application/dto"
	"user_service/internal/domain/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *dto.CreateUser) (*models.User, error)
	Follow(ctx context.Context, id, followerID string) error
	Unfollow(ctx context.Context, id, followerID string) error
}
