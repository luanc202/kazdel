package dto

import (
	"errors"
	"unicode"
)

type SignUpRequest struct {
	Name     string `form:"name"`
	Username string `form:"username"`
	Email    string `form:"email"`
	Password string `form:"password"`
	DTO
}

func NewSignUpRequest(username, email, password string) *SignUpRequest {
	return &SignUpRequest{
		Username: username,
		Email:    email,
		Password: password,
	}
}

func (s *SignUpRequest) Validate() error {
	if len(s.Name) < 1 || len(s.Name) > 20 {
		return errors.New("name must be between 1 and 20 characters")
	}

	if len(s.Username) < 4 || len(s.Username) > 20 {
		return errors.New("username must be between 4 and 20 characters")
	}

	if !usernameRegex.MatchString(s.Username) {
		return errors.New("username must contain only letters, numbers, dot and underscore")
	}

	if !emailRegex.MatchString(s.Email) && len(s.Email) > 2 && len(s.Email) < 255 {
		return errors.New("email must be a valid email address")
	}

	if len(s.Password) < 8 || len(s.Password) > 20 {
		return errors.New("password must be between 8 and 20 characters")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, c := range s.Password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsNumber(c):
			hasNumber = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return errors.New("password must include uppercase, lowercase, numbers and special characters")
	}

	return nil
}
