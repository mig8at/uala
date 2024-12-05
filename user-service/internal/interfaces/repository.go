package interfaces

import (
	"context"
	"user_service/internal/application/dto"
	"user_service/internal/domain/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *dto.CreateUser) (*models.User, error)
	GetById(ctx context.Context, id string) (*models.User, error)
	Paginate(ctx context.Context, page, limit int) ([]models.User, error)

	Follow(ctx context.Context, id, followerID string) error
	Unfollow(ctx context.Context, id, followerID string) error
	Followers(ctx context.Context, id string, page, limit int) ([]models.User, error)
	Following(ctx context.Context, id string, page, limit int) ([]models.User, error)
}
