package interfaces

import (
	"context"
	"tweet-service/internal/application/dto"
	"tweet-service/internal/domain/models"
)

type TweetRepository interface {
	Create(ctx context.Context, tweet *dto.CreateTweet) (*models.Tweet, error)
	Delete(ctx context.Context, id string) error
}
