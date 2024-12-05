package repository

import (
	"context"
	"encoding/json"
	"fmt"
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

	// Obtener los IDs de los tweets desde Redis
	tweetIDs, err := r.redis.LRange(ctx, timelineKey, int64(start), int64(end)).Result()
	if err != nil {
		return nil, fmt.Errorf("error al recuperar el timeline: %w", err)
	}

	if len(tweetIDs) == 0 {
		// No hay tweets para procesar
		return []*models.Timeline{}, nil
	}

	// Construir las claves de los tweets
	tweetKeys := make([]string, len(tweetIDs))
	for i, tweetID := range tweetIDs {
		tweetKeys[i] = fmt.Sprintf("tweets:%s", tweetID)
	}

	// Obtener todos los tweets de una vez
	tweetDataList, err := r.redis.MGet(ctx, tweetKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("error al recuperar los tweets: %w", err)
	}

	// Map para mantener los IDs de usuario únicos
	userIDSet := make(map[string]struct{})
	// Slice para mantener los tweets
	tweets := make([]*models.Tweet, 0, len(tweetDataList))

	for i, tweetData := range tweetDataList {
		if tweetData == nil {
			// El tweet no existe
			continue
		}
		tweetJSON, ok := tweetData.(string)
		if !ok {
			return nil, fmt.Errorf("tipo de dato inesperado para tweetID %s", tweetIDs[i])
		}

		var tweet models.Tweet
		if err := json.Unmarshal([]byte(tweetJSON), &tweet); err != nil {
			return nil, fmt.Errorf("error al deserializar tweetID %s: %w", tweetIDs[i], err)
		}
		tweet.ID = tweetIDs[i]
		tweets = append(tweets, &tweet)
		userIDSet[tweet.UserID] = struct{}{}
	}

	// Recopilar los IDs de usuario únicos
	userIDs := make([]string, 0, len(userIDSet))
	for uid := range userIDSet {
		userIDs = append(userIDs, uid)
	}

	// Construir las claves de los usuarios
	userKeys := make([]string, len(userIDs))
	for i, uid := range userIDs {
		userKeys[i] = fmt.Sprintf("users:%s", uid)
	}

	// Obtener todos los usuarios de una vez
	userDataList, err := r.redis.MGet(ctx, userKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("error al recuperar los usuarios: %w", err)
	}

	// Map para mantener los usuarios
	userMap := make(map[string]*models.User)
	for i, userData := range userDataList {
		if userData == nil {
			// El usuario no existe
			continue
		}
		userJSON, ok := userData.(string)
		if !ok {
			return nil, fmt.Errorf("tipo de dato inesperado para userID %s", userIDs[i])
		}

		var user models.User
		if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
			return nil, fmt.Errorf("error al deserializar userID %s: %w", userIDs[i], err)
		}

		userMap[userIDs[i]] = &user
	}

	// Construir el timeline
	timeline := make([]*models.Timeline, 0, len(tweets))
	for _, tweet := range tweets {
		user, ok := userMap[tweet.UserID]
		if !ok {
			// Usuario no encontrado, omitir este tweet
			continue
		}
		timeline = append(timeline, &models.Timeline{
			ID:       tweet.ID,
			Content:  tweet.Content,
			Likes:    tweet.Likes,
			Shares:   tweet.Shares,
			Comments: tweet.Comments,
			UserID:   tweet.UserID,
			Name:     user.Name,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		})
	}

	return timeline, nil
}
