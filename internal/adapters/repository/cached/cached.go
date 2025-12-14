package cached

import (
	"fmt"
	"log"

	"github.com/diogomassis/url-shortener/internal/core/domain"
	"github.com/diogomassis/url-shortener/internal/core/ports"
)

type cachedRepository struct {
	cache      ports.URLRepository
	persistent ports.URLRepository
}

func NewCachedRepository(cache ports.URLRepository, persistent ports.URLRepository) ports.URLRepository {
	return &cachedRepository{
		cache:      cache,
		persistent: persistent,
	}
}

func (r *cachedRepository) Save(url domain.URL) error {
	if err := r.persistent.Save(url); err != nil {
		return err
	}

	go func() {
		if err := r.cache.Save(url); err != nil {
			log.Printf("Failed to save to cache: %v", err)
		}
	}()

	return nil
}

func (r *cachedRepository) Get(shortCode string) (domain.URL, error) {
	url, err := r.cache.Get(shortCode)
	if err == nil {
		return url, nil
	}

	fmt.Printf("Cache miss for %s: %v\n", shortCode, err)
	url, err = r.persistent.Get(shortCode)
	if err != nil {
		return domain.URL{}, err
	}

	go func() {
		if err := r.cache.Save(url); err != nil {
			log.Printf("Failed to populate cache: %v", err)
		}
	}()

	return url, nil
}

func (r *cachedRepository) IncrementAccessCount(shortCode string) error {
	if err := r.persistent.IncrementAccessCount(shortCode); err != nil {
		return err
	}

	go func() {
		if err := r.cache.IncrementAccessCount(shortCode); err != nil {
			log.Printf("Failed to update cache access count: %v", err)
		}
	}()

	return nil
}
