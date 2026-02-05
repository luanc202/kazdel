package usecase

import (
	"fmt"
	"time"
	"url-shortener/m/entity"
	interfaces "url-shortener/m/interface"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	Repo      interfaces.UserRepository
	JwtSecret string
}

func NewAuthUseCase(repo interfaces.UserRepository, jwtSecret string) *AuthUseCase {
	return &AuthUseCase{
		Repo:      repo,
		JwtSecret: jwtSecret,
	}
}

func (uc *AuthUseCase) Signup(name, email, password string) error {
	existingUser, _ := uc.Repo.FindByEmail(email)
	if existingUser != nil {
		return fmt.Errorf("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := entity.NewUser(name, email, string(hashedPassword))

	err = uc.Repo.Save(user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (uc *AuthUseCase) Login(email, password string) (string, error) {
	user, err := uc.Repo.FindByEmail(email)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(uc.JwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

func (uc *AuthUseCase) GetUserByID(id string) (*entity.User, error) {
	// This might be needed later for user details, for now we just verify token validity in middleware
	return nil, nil
}
