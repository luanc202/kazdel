package interfaces

import (
	"url-shortener/m/entity"
)

type ShortenedUrlRepository interface {
	FindBySlug(slug string) (*entity.ShortenedUrl, error)

	Save(shortenedUrl *entity.ShortenedUrl) error
}
