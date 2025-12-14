package services

import (
	"math/rand"
	"time"

	"github.com/diogomassis/url-shortener/internal/core/domain"
	"github.com/diogomassis/url-shortener/internal/core/ports"
	"github.com/google/uuid"
)

type urlService struct {
	repo ports.URLRepository
}

func NewURLService(repo ports.URLRepository) ports.URLService {
	return &urlService{
		repo: repo,
	}
}

func (s *urlService) Shorten(originalURL string) (domain.URL, error) {
	shortCode := generateShortCode()
	url := domain.URL{
		ID:          uuid.New().String(),
		OriginalURL: originalURL,
		ShortCode:   shortCode,
		CreatedAt:   time.Now(),
		AccessCount: 0,
	}

	if err := s.repo.Save(url); err != nil {
		return domain.URL{}, err
	}

	return url, nil
}

func (s *urlService) GetOriginalURL(shortCode string) (string, error) {
	url, err := s.repo.Get(shortCode)
	if err != nil {
		return "", err
	}

	// Async increment to not block the response
	go s.repo.IncrementAccessCount(shortCode)

	return url.OriginalURL, nil
}

func generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
