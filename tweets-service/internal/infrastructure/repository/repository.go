package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"tweet-service/internal/application/dto"
	"tweet-service/internal/domain/models"
	"tweet-service/internal/interfaces"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type repository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewRepository(db *gorm.DB, redis *redis.Client) interfaces.TweetRepository {
	return &repository{db: db, redis: redis}
}

func (r *repository) Create(ctx context.Context, createTweetDTO *dto.CreateTweet) (*models.Tweet, error) {
	var tweet *models.Tweet

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Asignación explícita de campos
		tweet = &models.Tweet{
			Content: createTweetDTO.Content,
			UserID:  createTweetDTO.UserID,
			// Otros campos necesarios...
		}

		// Crear el tweet en la base de datos
		if err := tx.Create(tweet).Error; err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("operación cancelada por exceder el límite de tiempo")
			}
			return fmt.Errorf("error al crear el tweet: %w", err)
		}

		// Manejo de tags
		if len(createTweetDTO.Tags) > 0 {
			var tagModels []*models.Tag
			for _, tag := range createTweetDTO.Tags {
				tagName := cleanSpaces(tag)
				tagModel := &models.Tag{}

				// Buscar o crear el tag
				if err := tx.Where("name = ?", tagName).FirstOrCreate(tagModel, &models.Tag{Name: tagName}).Error; err != nil {
					return fmt.Errorf("error al crear o encontrar el tag '%s': %w", tagName, err)
				}

				tagModels = append(tagModels, tagModel)
			}

			// Asociar tags al tweet
			if err := tx.Model(tweet).Association("Tags").Append(tagModels); err != nil {
				return fmt.Errorf("error al asociar tags con el tweet: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if err := r.cacheTweet(ctx, tweet); err != nil {
		return nil, err
	}

	return tweet, nil
}

func (r *repository) cacheTweet(ctx context.Context, tweet *models.Tweet) error {
	tweetKey := fmt.Sprintf("tweets:%s", tweet.ID)

	tweetData, err := newTweet(tweet)
	if err != nil {
		return fmt.Errorf("error al serializar el tweet a JSON: %w", err)
	}

	pipe := r.redis.Pipeline()
	pipe.Set(ctx, tweetKey, tweetData, 0)

	if err := pipe.LPush(ctx, "tweet_queue", tweet.ID).Err(); err != nil {
		return fmt.Errorf("error al agregar el tweet a la cola: %w", err)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("error al ejecutar pipeline de Redis: %w", err)
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	tweet := &models.Tweet{}
	if err := r.db.WithContext(ctx).First(tweet, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("tweet no encontrado")
		}
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("operación cancelada por exceder el límite de tiempo")
		}
		return fmt.Errorf("error al obtener el tweet: %w", err)
	}

	// Iniciar transacción
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Eliminar el tweet de la base de datos
		if err := tx.Delete(tweet).Error; err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("operación cancelada por exceder el límite de tiempo")
			}
			return fmt.Errorf("error al eliminar el tweet: %w", err)
		}

		// Eliminar asociaciones si es necesario
		if err := tx.Model(tweet).Association("Tags").Clear(); err != nil {
			return fmt.Errorf("error al eliminar asociaciones de tags: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Eliminar el tweet de Redis
	tweetKey := fmt.Sprintf("tweet:%s", tweet.ID)
	timelineKey := fmt.Sprintf("timeline:%s", tweet.UserID)

	// Utilizar Pipeline para agrupar operaciones
	pipe := r.redis.Pipeline()
	pipe.Del(ctx, tweetKey)
	pipe.ZRem(ctx, timelineKey, tweet.ID)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("error al eliminar el tweet de Redis: %w", err)
	}

	return nil
}

func cleanSpaces(input string) string {
	trimmed := strings.TrimSpace(input)
	words := strings.Fields(trimmed)
	cleaned := strings.Join(words, " ")
	return cleaned
}

func newTweet(tw *models.Tweet) ([]byte, error) {
	jsonData, err := json.Marshal(struct {
		ID       string `json:"id"`
		UserID   string `json:"userId"`
		Content  string `json:"content"`
		Likes    int    `json:"likes"`
		Shares   int    `json:"shares"`
		Comments int    `json:"comments"`
	}{
		ID:       tw.ID,
		UserID:   tw.UserID,
		Content:  tw.Content,
		Likes:    tw.Likes,
		Shares:   tw.Shares,
		Comments: tw.CountComments,
	})
	if err != nil {
		return nil, fmt.Errorf("error al serializar el tweet a JSON: %w", err)
	}
	return jsonData, nil
}
