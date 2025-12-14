package ports

import "github.com/diogomassis/url-shortener/internal/core/domain"

type URLRepository interface {
	Save(url domain.URL) error
	Get(shortCode string) (domain.URL, error)
	IncrementAccessCount(shortCode string) error
}

type URLService interface {
	Shorten(originalURL string) (domain.URL, error)
	GetOriginalURL(shortCode string) (string, error)
}
