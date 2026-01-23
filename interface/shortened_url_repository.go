package interfaces

import (
	"url-shortener/m/entity"
)

type ShortenedUrlRepository interface {
	FindByID(id uint64) *entity.ShortenedUrl

	Save(shortenedUrl *entity.ShortenedUrl) error
	Update(shortenedUrlId uint64, shortenedUrlToUpdate *entity.ShortenedUrl) error
}
