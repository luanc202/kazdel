package entity

import (
	"time"
	"url-shortener/m/pkg/uniqueEntityId"
)

type ShortenedUrl struct {
	ID        uint64
	ShortSlug string
	LongUrl   string
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
	UserId    uniqueEntityId.ID
}

func NewShortenedUrl(shortSlug, longUrl string, expiresAt time.Time, userId uniqueEntityId.ID) *ShortenedUrl {
	return &ShortenedUrl{
		ShortSlug: shortSlug,
		LongUrl:   longUrl,
		ExpiresAt: expiresAt,
		UserId:    userId,
	}
}
