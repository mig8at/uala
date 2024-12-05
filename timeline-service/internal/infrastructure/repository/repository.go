package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"timeline-service/internal/domain/models"
	"timeline-service/internal/interfaces"

	"github.com/redis/go-redis/v9"
)

type Repository struct {
	redis *redis.Client
}

func NewRepository(redis *redis.Client) interfaces.Repository {
	return &Repository{redis: redis}
}

func (r *Repository) Paginate(ctx context.Context, userID string, page, size int) ([]*models.Timeline, error) {

	start := (page - 1) * size
	end := start + size - 1

	timelineKey := fmt.Sprintf("timeline:%s", userID)

	// Get the tweet IDs from Redis
	tweetIDs, err := r.redis.LRange(ctx, timelineKey, int64(start), int64(end)).Result()
	if err != nil {
		return nil, fmt.Errorf("error retrieving timeline: %w", err)
	}

	if len(tweetIDs) == 0 {
		// No tweets to process
		return []*models.Timeline{}, nil
	}

	// Build tweet keys
	tweetKeys := make([]string, len(tweetIDs))
	for i, tweetID := range tweetIDs {
		tweetKeys[i] = fmt.Sprintf("tweets:%s", tweetID)
	}

	// Fetch all tweets at once
	tweetDataList, err := r.redis.MGet(ctx, tweetKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("error retrieving tweets: %w", err)
	}

	// Map to hold unique user IDs
	userIDSet := make(map[string]struct{})
	// Slice to hold tweets
	tweets := make([]*models.Tweet, 0, len(tweetDataList))

	for i, tweetData := range tweetDataList {

		if tweetData == nil {
			// Tweet does not exist
			continue
		}
		tweetJSON, ok := tweetData.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected data type for tweetID %s", tweetIDs[i])
		}

		var tweet models.Tweet
		if err := json.Unmarshal([]byte(tweetJSON), &tweet); err != nil {
			return nil, fmt.Errorf("error unmarshalling tweetID %s: %w", tweetIDs[i], err)
		}
		tKey := tweetIDs[i]
		tweet.ID = tKey
		tweets = append(tweets, &tweet)
		userIDSet[tweet.UserID] = struct{}{}
	}

	fmt.Println(start, end, tweets, userIDSet)

	// Collect unique user IDs
	userIDs := make([]string, 0, len(userIDSet))
	for uid := range userIDSet {
		userIDs = append(userIDs, uid)
	}

	// Build user keys
	userKeys := make([]string, len(userIDs))
	for i, uid := range userIDs {
		userKeys[i] = fmt.Sprintf("users:%s", uid)
	}

	// Fetch all users at once
	userDataList, err := r.redis.MGet(ctx, userKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("error retrieving users: %w", err)
	}

	// Map to hold users
	userMap := make(map[string]*models.User)
	for i, userData := range userDataList {
		if userData == nil {
			// User does not exist
			continue
		}
		userJSON, ok := userData.(string)

		if !ok {
			return nil, fmt.Errorf("unexpected data type for userID %s", userIDs[i])
		}

		var user models.User
		if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
			return nil, fmt.Errorf("error unmarshalling userID %s: %w", userIDs[i], err)
		}

		userMap[userIDs[i]] = &user
	}

	fmt.Println(userDataList, len(userMap))

	// Build the timeline
	timeline := make([]*models.Timeline, 0, len(tweets))
	for _, tweet := range tweets {
		user, ok := userMap[tweet.UserID]
		fmt.Println(user)
		if !ok {
			// User not found, skip this tweet
			continue
		}
		timeline = append(timeline, &models.Timeline{
			ID:       tweet.ID,
			Content:  tweet.Content,
			Likes:    tweet.Likes,
			Shares:   tweet.Shares,
			Comments: tweet.Comments,
			UserID:   user.ID,
			Name:     user.Name,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		})
	}

	return timeline, nil
}

func getKey(key string) string {
	parts := strings.Split(key, ":")
	if len(parts) > 1 {
		return parts[1]
	}
	return key
}
