package memory

import (
	"errors"
	"sync"

	"github.com/diogomassis/url-shortener/internal/core/domain"
)

type memoryRepository struct {
	urls map[string]domain.URL
	mu   sync.RWMutex
}

func NewMemoryRepository() *memoryRepository {
	return &memoryRepository{
		urls: make(map[string]domain.URL),
	}
}

func (r *memoryRepository) Save(url domain.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.urls[url.ShortCode] = url
	return nil
}

func (r *memoryRepository) Get(shortCode string) (domain.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	url, ok := r.urls[shortCode]
	if !ok {
		return domain.URL{}, errors.New("URL not found")
	}
	return url, nil
}

func (r *memoryRepository) IncrementAccessCount(shortCode string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if url, ok := r.urls[shortCode]; ok {
		url.AccessCount++
		r.urls[shortCode] = url
		return nil
	}
	return errors.New("URL not found")
}
