package dto

import (
	"errors"
	"net/url"
	"time"
)

type ShortenedUrlInsert struct {
	OriginalUrl     string    `form:"originalUrl"`
	ExpiresAt       string    `form:"expiresAt"`
	ParsedExpiresAt time.Time `form:"-"`
}

func NewShortenedUrlInsert(originalUrl string, expiresAt string) *ShortenedUrlInsert {
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

	if s.ExpiresAt == "" {
		return errors.New("Expiration date cannot be empty")
	}

	// HTML datetime-local format: YYYY-MM-DDThh:mm
	parsedTime, err := time.Parse("2006-01-02T15:04", s.ExpiresAt)
	if err != nil {
		return errors.New("Expiration date must be a valid date and time")
	}
	s.ParsedExpiresAt = parsedTime

	if s.ParsedExpiresAt.Compare(time.Now()) < 0 {
		return errors.New("Expiration date cannot be in the past")
	}

	return nil
}
