package dto

import (
	"errors"
	"time"
)

type ShortenedUrlInsert struct {
	OriginalUrl string    `json:"originalUrl"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

func (s *ShortenedUrlInsert) Validate() error {
	if s.OriginalUrl == "" {
		return errors.New("Origin URL cannot be empty")
	}

	if s.ExpiresAt.Compare(time.Now()) < 0 {
		return errors.New("Expiration date cannot be in the past")
	}

	return nil
}
