package usecase

import (
	"fmt"
	"kazdel/pkg/entity"
	interfaces "kazdel/pkg/interface"
	"kazdel/pkg/uniqueEntityId"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	UserRepo    interfaces.UserRepository
	SessionRepo interfaces.SessionRepository
}

func NewAuthUseCase(userRepo interfaces.UserRepository, sessionRepo interfaces.SessionRepository) *AuthUseCase {
	return &AuthUseCase{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
	}
}

func (uc *AuthUseCase) Signup(name, username, email, password string) (string, error) {
	exists, _ := uc.UserRepo.ExistsByEmail(email)
	if exists {
		return "", fmt.Errorf("email already registered")
	}

	exists, _ = uc.UserRepo.ExistsByUsername(username)
	if exists {
		return "", fmt.Errorf("username already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("failed to hash password", "error", err)
		return "", fmt.Errorf("failed to create user")
	}

	user := entity.NewUser(name, username, entity.RoleUser, email, string(hashedPassword))

	err = uc.UserRepo.Save(user)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	token := uniqueEntityId.NewID().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	session := entity.NewSession(user.ID, token, expiresAt)

	err = uc.SessionRepo.Create(session)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return token, nil
}

func (uc *AuthUseCase) Login(email, password string) (string, error) {
	user, err := uc.UserRepo.FindByEmail(email)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	// Generate a session token (using UUID for simplicity and uniqueness)
	token := uniqueEntityId.NewID().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	session := entity.NewSession(user.ID, token, expiresAt)

	err = uc.SessionRepo.Create(session)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return token, nil
}

func (uc *AuthUseCase) Logout(token string) error {
	return uc.SessionRepo.DeleteByToken(token)
}

func (uc *AuthUseCase) ValidateSession(token string) (string, error) {
	session, err := uc.SessionRepo.FindByToken(token)
	if err != nil {
		return "", err
	}
	if session == nil {
		return "", fmt.Errorf("session not found")
	}

	if session.ExpiresAt.Before(time.Now()) {
		uc.SessionRepo.DeleteByToken(token) // Clean up expired session
		return "", fmt.Errorf("session expired")
	}

	return session.UserID.String(), nil
}

func (uc *AuthUseCase) GetUserByID(id string) (*entity.User, error) {
	// Not strictly needed for auth flow but good to keep
	// ... implementation if needed or remove ...
	return nil, nil
}
