package dto

import (
	"errors"
)

type LoginRequest struct {
	Username string `form:"username"`
	Password string `form:"password"`
	DTO
}

func NewLoginRequest(username, password string) *LoginRequest {
	return &LoginRequest{
		Username: username,
		Password: password,
	}
}

func (l *LoginRequest) Validate() error {
	if l.Username == "" || l.Password == "" {
		return errors.New("username and password are required")
	}

	if !usernameRegex.MatchString(l.Username) {
		return errors.New("username must contain only letters, numbers, underscores and dots")
	}

	if len(l.Username) < 4 || len(l.Username) > 20 {
		return errors.New("username must be between 4 and 20 characters long")
	}

	if len(l.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	return nil
}
