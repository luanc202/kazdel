package interfaces

import (
	"time"

	"kazdel/pkg/entity"
	"kazdel/pkg/uniqueEntityId"
)

type ShortenedUrlRepository interface {
	FindBySlug(slug string) (*entity.ShortenedUrl, error)

	FindByUserIdPaginated(userId uniqueEntityId.ID, search string, page, limit int) ([]*entity.ShortenedUrl, int, error)

	Save(shortenedUrl *entity.ShortenedUrl) error
	Update(shortenedUrl *entity.ShortenedUrl) error
	Delete(id uint64, userId uniqueEntityId.ID) error
	DeleteExpired(now time.Time) error
}
