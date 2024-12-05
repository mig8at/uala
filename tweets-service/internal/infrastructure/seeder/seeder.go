package seeder

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
	"tweet-service/internal/application/dto"
	"tweet-service/internal/interfaces"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Seeder struct {
	db    *gorm.DB
	redis *redis.Client
	repo  interfaces.TweetRepository
}

func NewSeeder(db *gorm.DB, redis *redis.Client, repo interfaces.TweetRepository) *Seeder {
	return &Seeder{db: db, redis: redis, repo: repo}
}

func generateRandomContent(rnd *rand.Rand) string {
	phrases := []string{
		"Hello, World!", "Exploring the wonders of Go!", "Cloud development is amazing!",
		"Writing clean code is an art.", "Have you tried learning a new language?",
		"Backend development is my passion.", "Debugging is like solving a mystery.",
		"Unit tests are essential for good software.", "Open source contributions are fulfilling.",
		"Always remember to document your code.", "Coding marathons are intense but fun!",
		"The future is in the cloud.", "Embrace the challenges of learning new tech.",
		"Startups need great software architects.", "APIs are the backbone of modern software.",
		"The beauty of algorithms is unmatched.", "Write once, run anywhere!",
		"Debugging can be frustrating but rewarding.", "Why is Go so fast? Explore and learn.",
		"Let's simplify things with microservices.",
	}
	content := []string{}

	for len(strings.Join(content, " ")) < 280 {
		phrase := phrases[rnd.Intn(len(phrases))]
		if len(strings.Join(content, " ")+phrase) > 280 {
			break
		}
		content = append(content, phrase)
	}

	return strings.Join(content, " ")
}

func getRandomTags(pool []string, count int) []string {
	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	if count > len(pool) {
		count = len(pool)
	}
	return pool[:count]
}

func generateTweets() []dto.CreateTweet {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	var tweets []dto.CreateTweet

	idsUsers := []string{
		"2a42c7ae-7f78-4e36-8358-902342fe23f1",
		"83836283-0760-4879-a7df-af4769a2d1a4",
		"2327a87b-3fe7-4bc9-a275-75d33358f1bc",
		"26474281-d97a-474d-a593-68aa1c1f48ef",
		"2e3b1b92-62ba-4308-8872-6a3d964f3a80",
		"444f79b1-c805-4998-a4db-24086790b031",
		"4774f12c-8c0c-4bfc-9692-f2b220553023",
		"4eef46a3-fe6d-4a60-af4c-fd354f987cc8",
		"5f5d1800-e9a2-496f-8cae-52ea1f587acc",
		"65c44c69-8663-4e00-ad7c-90fd48e95102",
		"6cebd913-085d-4144-a946-3d8fdfadae36",
		"7f35a8d8-7af5-4e4d-a26c-e0afe5245bca",
		"812f2527-cb7d-4b8a-9899-8035c301cf29",
		"8372ea9d-2424-4650-b10b-9f9a15a50a6e",
		"c2dae4b6-d8ba-4b4c-89ef-39e0d92df28e",
		"d6d6a128-ad31-45bd-8758-7bddd9871a05",
		"e1dfb9c5-5fb1-4157-809f-4c78c9b5d355",
		"e574036a-4027-4dca-b698-6f9be442b03f",
		"f3335e2a-d681-4e2a-9024-1164c85c5f87",
		"fbe5a0a7-8ecb-4868-a9a8-54a280cd8edc",
	}

	tagPool := []string{"golang", "programming", "tech", "devlife", "coding", "backend", "cloud", "webdev", "API", "openSource"}

	for i := 0; i < 500; i++ {
		tagCount := rnd.Intn(5) + 1 // Al menos 1 tag
		newTags := getRandomTags(tagPool, tagCount)

		userID := idsUsers[rnd.Intn(len(idsUsers))]

		tweet := dto.CreateTweet{
			UserID:  userID,
			Content: generateRandomContent(rnd),
			Tags:    newTags,
		}

		tweets = append(tweets, tweet)
	}

	return tweets
}

func (s *Seeder) Seed() {
	s.Clean()

	tweets := generateTweets()

	for _, tweet := range tweets {
		s.repo.Create(context.Background(), &tweet)
	}
}

func (s *Seeder) Clean() {
	ctx := context.Background()
	for _, table := range []string{"tweet_tags", "comments", "tags", "tweets"} {
		// Eliminar contenido de cada tabla
		err := s.db.Exec("DELETE FROM " + table).Error
		if err != nil {
			log.Fatalf("Error al borrar el contenido de la tabla %s: %v", table, err)
		}
	}

	if err := deleteKeysWithPrefix(ctx, s.redis, "tweets:"); err != nil {
		fmt.Printf("Error al borrar claves con prefijo tweets:: %v\n", err)
	}
}

func deleteKeysWithPrefix(ctx context.Context, rdb *redis.Client, prefix string) error {
	// Usamos SCAN para buscar claves que coincidan con el prefijo
	iter := rdb.Scan(ctx, 0, prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		err := rdb.Del(ctx, key).Err()
		if err != nil {
			return fmt.Errorf("error al borrar la clave %s: %w", key, err)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("error durante la iteración de SCAN: %w", err)
	}

	return nil
}
