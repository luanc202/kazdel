package entity

import (
	"kazdel/pkg/uniqueEntityId"
	"time"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID           uniqueEntityId.ID
	Name         string
	Username     string
	Role         Role
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(name, username string, role Role, email, passwordHash string) *User {
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
