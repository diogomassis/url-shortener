package services

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/diogomassis/url-shortener/internal/core/domain"
	"github.com/diogomassis/url-shortener/internal/core/ports"
	"github.com/google/uuid"
	"github.com/speps/go-hashids/v2"
)

const (
	// MaxRetries is the maximum number of times to retry generating a unique short code.
	MaxRetries = 10
	// ShortCodeLength is the desired length of the short code.
	ShortCodeLength = 7
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
	for i := 0; i < MaxRetries; i++ {
		shortCode, err := s.generateUniqueShortCode(originalURL, i)
		if err != nil {
			return domain.URL{}, fmt.Errorf("failed to generate short code: %w", err)
		}

		// Check for collision
		if _, err := s.repo.Get(shortCode); err != nil {
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
	}

	return domain.URL{}, fmt.Errorf("failed to generate a unique short code after %d retries", MaxRetries)
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

func (s *urlService) generateUniqueShortCode(originalURL string, salt int) (string, error) {
	// 1. Create a deterministic hash from the URL + salt
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s-%d", originalURL, salt)))
	hash := hasher.Sum(nil)

	// 2. Use Hashids to create a short, unique, non-sequential ID from the hash
	hd := hashids.NewData()
	hd.Salt = "this is my salt" // Use a secret salt
	hd.MinLength = ShortCodeLength
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", err
	}

	// Convert the first few bytes of the hash to a number to be encoded
	var numbers []int
	for i := 0; i < 4; i++ {
		numbers = append(numbers, int(hash[i]))
	}

	encoded, err := h.Encode(numbers)
	if err != nil {
		return "", err
	}

	// 3. Fallback to a simpler encoding if hashids is too long (it shouldn't be with this setup)
	if len(encoded) > ShortCodeLength {
		encoded = base64.URLEncoding.EncodeToString(hash)[:ShortCodeLength]
	}

	return encoded, nil
}
