package entity

import (
	"time"
	"kazdel/pkg/uniqueEntityId"
)

type Session struct {
	ID        uniqueEntityId.ID
	UserID    uniqueEntityId.ID
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewSession(userID uniqueEntityId.ID, token string, expiresAt time.Time) *Session {
	return &Session{
		ID:        uniqueEntityId.NewID(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
}
