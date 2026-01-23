package entity

import (
	"time"
	"url-shortener/m/internal/uniqueEntityId"
)

type ShortenedUrl struct {
	ID          uint64
	ShortSlug   string
	OriginalUrl string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ExpiresAt   time.Time
	UserId      uniqueEntityId.ID
}

func NewShortenedUrl(shortSlug, originalUrl string, expiresAt time.Time, userId uniqueEntityId.ID) *ShortenedUrl {
	return &ShortenedUrl{
		ShortSlug:   shortSlug,
		OriginalUrl: originalUrl,
		ExpiresAt:   expiresAt,
		UserId:      userId,
	}
}
