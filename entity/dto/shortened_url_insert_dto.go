package dto

import (
	"errors"
)

type ShortenedUrlInsertDto struct {
	OriginalUrl string `json:"originalUrl"`
	ExpiresAt   int    `json:"expiresAt"`
}

func (s *ShortenedUrlInsertDto) Validate() error {
	if s.OriginalUrl == "" {
		return errors.New("Original URL cannot be empty")
	}

	if s.ExpiresAt == 0 {
		return errors.New("Expires At cannot be empty")
	}

	return nil
}
