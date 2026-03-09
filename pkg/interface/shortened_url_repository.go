package interfaces

import (
	"url-shortener/m/entity"
	"url-shortener/m/internal/uniqueEntityId"
)

type ShortenedUrlRepository interface {
	FindBySlug(slug string) (*entity.ShortenedUrl, error)

	FindByUserId(userId uniqueEntityId.ID) ([]*entity.ShortenedUrl, error)

	Save(shortenedUrl *entity.ShortenedUrl) error
	Delete(id uint64, userId uniqueEntityId.ID) error
}
