package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"timeline-service/internal/domain/models"
	"timeline-service/internal/interfaces"

	"github.com/redis/go-redis/v9"
)

type cron struct {
	redis *redis.Client
}

func NewCron(redis *redis.Client) interfaces.Cron {
	return &cron{redis: redis}
}

func (c *cron) ProcessTweets() {
	ctx := context.Background()
	numWorkers := 5 // Número de trabajadores

	// Canal para comunicar los IDs de tweets
	tweetIDs := make(chan string)

	// Iniciar trabajadores
	for i := 0; i < numWorkers; i++ {
		go c.worker(ctx, tweetIDs)
	}

	luaScript := redis.NewScript(`
local tweet_id = redis.call('RPOP', KEYS[1])
if tweet_id then
    local exists = redis.call('SISMEMBER', KEYS[2], tweet_id)
    if exists == 0 then
        redis.call('SADD', KEYS[2], tweet_id)
        return tweet_id
    end
end
return nil
    `)

	for {
		result, err := luaScript.Run(ctx, c.redis, []string{"tweet_queue", "processed_tweets"}).Result()
		if err != nil && err != redis.Nil {
			log.Printf("Error al ejecutar el script de Lua: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if result == nil {
			time.Sleep(2 * time.Second)
			continue
		}

		tweetID, ok := result.(string)
		if !ok {
			log.Printf("Tipo de dato inesperado: %T", result)
			continue
		}

		// Enviar el tweetID al canal para que lo procesen los trabajadores
		tweetIDs <- tweetID
	}
}

func (c *cron) worker(_ context.Context, tweetIDs <-chan string) {
	for tweetID := range tweetIDs {
		if err := c.processTweet(tweetID); err != nil {
			log.Printf("Error al procesar el tweet ID=%s: %v", tweetID, err)
			// Manejar el error según sea necesario
		}
	}
}

func (c *cron) processTweet(tweetID string) error {
	ctx := context.Background()
	tweet, err := c.getTweet(ctx, tweetID)
	if err != nil {
		return fmt.Errorf("error al obtener el tweet: %w", err)
	}

	followers, err := c.getFollowers(ctx, tweet.UserID)
	if err != nil {
		return fmt.Errorf("error al obtener los seguidores: %w", err)
	}

	pipe := c.redis.Pipeline()
	for _, followerID := range followers {
		timelineKey := fmt.Sprintf("timeline:%s", followerID)
		pipe.LPush(ctx, timelineKey, tweetID)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("error al actualizar los timelines: %w", err)
	}

	fmt.Printf("Procesado tweet: ID=%s\n", tweet.ID)
	return nil
}

func (r *cron) getTweet(ctx context.Context, tweetID string) (*models.Tweet, error) {
	// Construye la clave del tweet
	tweetKey := fmt.Sprintf("tweets:%s", tweetID)

	// Obtén el valor de Redis
	tweetData, err := r.redis.Get(ctx, tweetKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("el tweet con ID %s no existe", tweetID)
		}
		return nil, fmt.Errorf("error al obtener el tweet: %w", err)
	}

	// Deserializa el JSON al objeto Tweet
	var tweet models.Tweet
	err = json.Unmarshal([]byte(tweetData), &tweet)
	if err != nil {
		return nil, fmt.Errorf("error al deserializar el JSON del tweet: %w", err)
	}

	return &tweet, nil
}

func (r *cron) getFollowers(ctx context.Context, userID string) ([]string, error) {
	key := fmt.Sprintf("followers:%s", userID)

	followers, err := r.redis.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("error al obtener los seguidores: %w", err)
	}

	return followers, nil
}
