package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"timeline-service/internal/domain/models"
	"timeline-service/internal/interfaces"

	"github.com/dgraph-io/badger/v4"
)

type Repository struct {
	db *badger.DB
}

func NewRepository(db *badger.DB) interfaces.Repository {
	return &Repository{db: db}
}

func (r *Repository) Paginate(ctx context.Context, userID string, limit, offset int) ([]*models.Tweet, error) {
	tweets := make([]*models.Tweet, 0)

	err := r.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = limit + offset
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("tweets:")
		count := 0
		skipped := 0

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if skipped < offset {
				skipped++
				continue
			}
			if count >= limit {
				break
			}

			item := it.Item()
			var tweet models.Tweet
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &tweet)
			})
			if err != nil {
				return fmt.Errorf("error deserializando el tweet: %w", err)
			}

			tweets = append(tweets, &tweet)
			count++
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return tweets, nil
}

func (r *Repository) SaveUser(ctx context.Context, user *models.User) error {
	key := []byte(fmt.Sprintf("users:%s", user.ID))

	// Serializa el usuario a JSON
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("error serializando el usuario: %w", err)
	}

	// Inicia una transacción de escritura
	err = r.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, data)
		if err != nil {
			return fmt.Errorf("error guardando el usuario: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) SaveTweet(ctx context.Context, tweet *models.Tweet) error {
	// Construye la clave
	key := []byte(fmt.Sprintf("tweets:%s:%s", tweet.UserID, tweet.ID))

	// Serializa el tweet a JSON
	data, err := json.Marshal(tweet)
	if err != nil {
		return fmt.Errorf("error serializando el tweet: %w", err)
	}

	// Inicia una transacción de escritura
	err = r.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, data)
		if err != nil {
			return fmt.Errorf("error guardando el tweet: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
