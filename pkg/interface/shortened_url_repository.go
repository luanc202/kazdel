package interfaces

import (
	"url-shortener/m/pkg/entity"
	"url-shortener/m/pkg/uniqueEntityId"
)

type ShortenedUrlRepository interface {
	FindBySlug(slug string) (*entity.ShortenedUrl, error)

	FindByUserId(userId uniqueEntityId.ID) ([]*entity.ShortenedUrl, error)

	Save(shortenedUrl *entity.ShortenedUrl) error
	Delete(id uint64, userId uniqueEntityId.ID) error
}
