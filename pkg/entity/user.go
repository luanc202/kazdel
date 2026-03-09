package entity

import (
	"time"
	"url-shortener/m/internal/uniqueEntityId"
)

type User struct {
	ID           uniqueEntityId.ID
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(name, email, passwordHash string) *User {
	return &User{
		ID:           uniqueEntityId.NewID(),
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}
