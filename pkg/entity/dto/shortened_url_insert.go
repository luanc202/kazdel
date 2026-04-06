package dto

import (
	"errors"
	"net/url"
	"time"
)

type ShortenedUrlInsert struct {
	OriginalUrl string    `json:"originalUrl"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

func NewShortenedUrlInsert(originalUrl string, expiresAt time.Time) *ShortenedUrlInsert {
	return &ShortenedUrlInsert{
		OriginalUrl: originalUrl,
		ExpiresAt:   expiresAt,
	}
}

func (s *ShortenedUrlInsert) Validate() error {
	if s.OriginalUrl == "" {
		return errors.New("Origin URL cannot be empty")
	}

	parsedUrl, err := url.ParseRequestURI(s.OriginalUrl)
	if err != nil || (parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https") || parsedUrl.Host == "" {
		return errors.New("Origin URL must be a valid HTTP or HTTPS URL")
	}

	if s.ExpiresAt.IsZero() {
		return errors.New("Expiration date cannot be empty")
	}

	if s.ExpiresAt.Compare(time.Now()) < 0 {
		return errors.New("Expiration date cannot be in the past")
	}

	return nil
}
