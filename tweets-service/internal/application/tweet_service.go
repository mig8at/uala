package application

import (
	"context"
	"time"
	"tweet-service/internal/application/dto"
	"tweet-service/internal/interfaces"

	"github.com/jinzhu/copier"
)

type tweetservice struct {
	repo interfaces.TweetRepository
}

func NewService(repo interfaces.TweetRepository) interfaces.Tweetservice {
	return &tweetservice{
		repo: repo,
	}
}

func (s *tweetservice) Create(ctx context.Context, tweet *dto.CreateTweet) (*dto.Tweet, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	newtweet, err := s.repo.Create(ctx, tweet)
	if err != nil {
		return nil, err
	}

	tweetDTO := &dto.Tweet{}
	if err := copier.Copy(tweetDTO, newtweet); err != nil {
		return nil, err
	}

	return tweetDTO, nil
}

func (s *tweetservice) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
