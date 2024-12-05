package application

import (
	"context"
	"time"
	"timeline-service/internal/domain/models"
	"timeline-service/internal/interfaces"
)

type tweetservice struct {
	repo   interfaces.Repository
	client interfaces.Client
}

func NewService(repo interfaces.Repository, client interfaces.Client) interfaces.Service {
	return &tweetservice{
		repo:   repo,
		client: client,
	}
}

func (s *tweetservice) Paginate(ctx context.Context, id string, limit, offset int) ([]*models.Tweet, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	tweets, err := s.repo.Paginate(ctx, id, limit, offset)
	if err != nil {
		return nil, err
	}

	return tweets, nil
}

func (s *tweetservice) SyncData(ctx context.Context) error {

	tweets, err := s.client.Tweets(ctx, 1, 10)
	if err != nil {
		return err
	}

	for _, tweet := range tweets {
		if err := s.repo.SaveTweet(ctx, tweet); err != nil {
			return err
		}

		user, err := s.client.User(ctx, tweet.UserID)
		if err != nil {
			return err
		}

		if err := s.repo.SaveUser(ctx, user); err != nil {
			return err
		}
	}

	return nil
}
