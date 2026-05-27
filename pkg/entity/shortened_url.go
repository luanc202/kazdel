package entity

import (
	"time"
	"kazdel/pkg/uniqueEntityId"
)

type ShortenedUrl struct {
	ID           uint64
	ShortSlug    string
	LongUrl      string
	Description  *string
	PasswordHash *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ExpiresAt    time.Time
	UserId       uniqueEntityId.ID
	Views        int64
}

func NewShortenedUrl(shortSlug, longUrl string, expiresAt time.Time, userId uniqueEntityId.ID, description *string, passwordHash *string) *ShortenedUrl {
	return &ShortenedUrl{
		ShortSlug:    shortSlug,
		LongUrl:      longUrl,
		Description:  description,
		PasswordHash: passwordHash,
		ExpiresAt:    expiresAt,
		UserId:       userId,
	}
}
