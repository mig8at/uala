package interfaces

import (
	"context"
	"tweet-service/internal/application/dto"
)

type Tweetservice interface {
	Create(ctx context.Context, tweet *dto.CreateTweet) (*dto.Tweet, error)
	Delete(ctx context.Context, id string) error
}
