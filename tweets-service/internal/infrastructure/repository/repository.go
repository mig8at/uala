package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"tweet-service/internal/application/dto"
	"tweet-service/internal/domain/models"
	"tweet-service/internal/interfaces"

	"github.com/jinzhu/copier"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type repository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewRepository(sqlite *gorm.DB, redis *redis.Client) interfaces.TweetRepository {
	return &repository{db: sqlite, redis: redis}
}

func (r *repository) Create(ctx context.Context, CreateTweet *dto.CreateTweet) (*models.Tweet, error) {
	var tweet models.Tweet

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := copier.Copy(&tweet, CreateTweet); err != nil {
			return fmt.Errorf("error al copiar datos del DTO al modelo Tweet: %w", err)
		}

		tweet.Tags = nil

		if err := tx.Create(&tweet).Error; err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("operación cancelada por exceder el límite de tiempo")
			}
			return fmt.Errorf("error al crear el tweet: %w", err)
		}

		for _, tag := range CreateTweet.Tags {
			tagModel := &models.Tag{}
			tagName := cleanSpaces(tag)

			result := tx.Where("name = ?", tagName).FirstOrCreate(tagModel, models.Tag{
				Name: tagName,
			})

			if result.Error != nil {
				return fmt.Errorf("error al crear o encontrar el tag '%s': %w", tagName, result.Error)
			}

			if err := tx.Model(&tweet).Association("Tags").Append(tagModel); err != nil {
				return fmt.Errorf("error al asociar el tag '%s' con el tweet: %w", tagName, err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Almacenar el tweet en Redis
	tweetKey := fmt.Sprintf("tweet:%s", tweet.ID)
	tweetData, err := newTweet(&tweet)
	if err != nil {
		return nil, fmt.Errorf("error al serializar el tweet a JSON: %w", err)
	}

	err = r.redis.Set(ctx, tweetKey, tweetData, 0).Err()
	if err != nil {
		fmt.Printf("Error al guardar el tweet en Redis: %v\n", err)
	}

	// Actualizar el timeline del usuario en Redis
	err = r.redis.ZAdd(ctx, fmt.Sprintf("tweets:%s", tweet.UserID), redis.Z{
		Score:  float64(tweet.CreatedAt.Unix()),
		Member: tweet.ID,
	}).Err()

	if err != nil {
		fmt.Printf("Error al guardar el tweet en el timeline del usuario en Redis: %v\n", err)
	}

	return &tweet, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {

	tweet := &models.Tweet{}
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(tweet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("tweet no encontrado")
		}
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("operación cancelada por exceder el límite de tiempo")
		}
		return err
	}

	if err := r.db.WithContext(ctx).Delete(tweet).Error; err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("operación cancelada por exceder el límite de tiempo")
		}
		return err
	}

	key := fmt.Sprintf("tweets:%s:%s", tweet.UserID, tweet.ID)
	if err := r.redis.Del(ctx, key).Err(); err != nil {
		fmt.Printf("Error al eliminar clave en Redis: %v\n", err)
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
		Content  string `json:"content"`
		Likes    int    `json:"likes"`
		Shares   int    `json:"shares"`
		Comments int    `json:"comments"`
	}{
		Content:  tw.Content,
		Likes:    tw.Likes,
		Shares:   tw.Shares,
		Comments: tw.CountComments,
	})
	if err != nil {
		return nil, fmt.Errorf("error al serializar el Tweet a JSON: %w", err)
	}
	return jsonData, nil
}
