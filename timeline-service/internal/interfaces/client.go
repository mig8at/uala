package interfaces

import (
	"context"
	"timeline-service/internal/domain/models"
)

type Client interface {
	User(ctx context.Context, userID string) (*models.User, error)
	Followers(ctx context.Context, userID string, page, limit int) ([]*models.User, error)
	Tweets(ctx context.Context, page, limit int) ([]*models.Tweet, error)
}
