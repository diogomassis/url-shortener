package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/diogomassis/url-shortener/internal/core/domain"
	"github.com/redis/go-redis/v9"
)

type redisRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisRepository(addr string, password string, db int, ttl time.Duration) *redisRepository {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &redisRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *redisRepository) Save(url domain.URL) error {
	ctx := context.Background()
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, url.ShortCode, data, r.ttl).Err()
}

func (r *redisRepository) Get(shortCode string) (domain.URL, error) {
	ctx := context.Background()
	val, err := r.client.Get(ctx, shortCode).Result()
	if err != nil {
		if err == redis.Nil {
			return domain.URL{}, errors.New("URL not found in cache")
		}
		return domain.URL{}, err
	}

	var url domain.URL
	if err := json.Unmarshal([]byte(val), &url); err != nil {
		return domain.URL{}, err
	}

	return url, nil
}

func (r *redisRepository) IncrementAccessCount(shortCode string) error {
	// For cache, we might just want to update the value if it exists.
	// But since we store the whole object as JSON, we need to Get, Update, Set.
	// Or we can just ignore increment on cache if it's just a cache.
	// However, to keep it consistent, let's try to update it.
	// A better approach for Redis would be storing fields in a Hash, but let's stick to simple JSON for now.

	// Optimistic locking or just simple get-set for this example.
	url, err := r.Get(shortCode)
	if err != nil {
		return err
	}

	url.AccessCount++
	return r.Save(url)
}
