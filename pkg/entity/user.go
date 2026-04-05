package entity

import (
	"kazdel/pkg/uniqueEntityId"
	"time"
)

type User struct {
	ID           uniqueEntityId.ID
	Name         string
	Username     string
	Role         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(name, username, role, email, passwordHash string) *User {
	return &User{
		ID:           uniqueEntityId.NewID(),
		Name:         name,
		Username:     username,
		Role:         role,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}
