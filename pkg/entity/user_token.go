package entity

import (
	"kazdel/pkg/uniqueEntityId"
	"time"
)

type TokenContext string

const (
	TokenContextEmailVerification TokenContext = "email_verification"
	TokenContextPasswordReset     TokenContext = "password_reset"
)

type UserToken struct {
	ID        uniqueEntityId.ID
	UserID    uniqueEntityId.ID
	Token     string
	Context   TokenContext
	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewUserToken(userID uniqueEntityId.ID, token string, context TokenContext, expiresIn time.Duration) *UserToken {
	return &UserToken{
		ID:        uniqueEntityId.NewID(),
		UserID:    userID,
		Token:     token,
		Context:   context,
		ExpiresAt: time.Now().Add(expiresIn),
		CreatedAt: time.Now(),
	}
}

func (t *UserToken) IsExpired() bool {
	return t.ExpiresAt.Before(time.Now())
}
