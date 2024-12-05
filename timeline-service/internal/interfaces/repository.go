package interfaces

import (
	"context"
	"timeline-service/internal/domain/models"
)

type Repository interface {
	Paginate(ctx context.Context, id string, limit, offset int) ([]*models.Tweet, error)
	SaveTweet(ctx context.Context, tweet *models.Tweet) error
	SaveUser(ctx context.Context, user *models.User) error
}
