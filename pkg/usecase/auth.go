package usecase

import (
	"errors"
	"fmt"
	"kazdel/pkg/entity"
	"kazdel/pkg/infra/config"
	interfaces "kazdel/pkg/interface"
	"kazdel/pkg/uniqueEntityId"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmailNotVerified = errors.New("email_not_verified")

type AuthUseCase struct {
	UserRepo      interfaces.UserRepository
	SessionRepo   interfaces.SessionRepository
	UserTokenRepo interfaces.UserTokenRepository
	EmailService  interfaces.EmailService
}

func NewAuthUseCase(userRepo interfaces.UserRepository, sessionRepo interfaces.SessionRepository, userTokenRepo interfaces.UserTokenRepository, emailService interfaces.EmailService) *AuthUseCase {
	return &AuthUseCase{
		UserRepo:      userRepo,
		SessionRepo:   sessionRepo,
		UserTokenRepo: userTokenRepo,
		EmailService:  emailService,
	}
}

func (uc *AuthUseCase) Signup(name, username, email, password string) (string, error) {
	exists, _ := uc.UserRepo.ExistsByEmail(email)
	if exists {
		return "", fmt.Errorf("Email already registered")
	}

	exists, _ = uc.UserRepo.ExistsByUsername(username)
	if exists {
		return "", fmt.Errorf("Username already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		return "", fmt.Errorf("Failed to create user")
	}

	user := entity.NewUser(name, username, entity.RoleUser, email, string(hashedPassword))

	err = uc.UserRepo.Save(user)
	if err != nil {
		slog.Error("Failed to save user", "error", err)
		return "", fmt.Errorf("Failed to create user")
	}

	token := uniqueEntityId.NewID().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	session := entity.NewSession(user.ID, token, expiresAt)

	err = uc.SessionRepo.Create(session)
	if err != nil {
		return "", fmt.Errorf("Failed to create session: %w", err)
	}

	// Generate email verification token
	verificationToken := uniqueEntityId.NewID().String()
	userToken := entity.NewUserToken(user.ID, verificationToken, entity.TokenContextEmailVerification, 24*time.Hour)
	err = uc.UserTokenRepo.Save(userToken)
	if err != nil {
		slog.Error("Failed to save user verification token", "error", err)
	} else {
		// e.g. domain.com/api/v1/auth/verify-email?token=xxx
		// We'll pass a relative URL or get the base URL from env if needed.
		// For simplicity, passing /auth/verify-email?token=xxx
		verifyLink := fmt.Sprintf("/auth/verify-email?token=%s", verificationToken)
		uc.EmailService.SendVerificationEmail(user.Email, user.Name, verifyLink)
	}

	return token, nil
}

func (uc *AuthUseCase) Login(username, password string) (string, error) {
	user, err := uc.UserRepo.FindByUsername(username)
	if err != nil {
		return "", fmt.Errorf("Invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", fmt.Errorf("Invalid credentials")
	}

	env := config.GetEnvConfig()
	if env != nil && env.MAIL_ENABLED && !user.EmailVerified {
		return "", ErrEmailNotVerified
	}

	// Generate a session token (using UUID for simplicity and uniqueness)
	token := uniqueEntityId.NewID().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	session := entity.NewSession(user.ID, token, expiresAt)

	err = uc.SessionRepo.Create(session)
	if err != nil {
		slog.Error("Failed to create session", "error", err)
		return "", fmt.Errorf("Failed to create session")
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
		return "", fmt.Errorf("Session not found")
	}

	if session.ExpiresAt.Before(time.Now()) {
		uc.SessionRepo.DeleteByToken(token) // Clean up expired session
		return "", fmt.Errorf("Session expired")
	}

	return session.UserID.String(), nil
}

func (uc *AuthUseCase) GetUserByID(id string) (*entity.User, error) {
	return uc.UserRepo.FindById(id)
}

func (uc *AuthUseCase) GetUserByUsername(username string) (*entity.User, error) {
	return uc.UserRepo.FindByUsername(username)
}

func (uc *AuthUseCase) VerifyEmail(token string) error {
	userToken, err := uc.UserTokenRepo.FindByToken(token)
	if err != nil {
		return fmt.Errorf("Invalid or expired token")
	}

	if userToken.Context != entity.TokenContextEmailVerification || userToken.IsExpired() {
		return fmt.Errorf("Invalid or expired token")
	}

	user, err := uc.UserRepo.FindById(userToken.UserID.String())
	if err != nil {
		return fmt.Errorf("User not found")
	}

	user.EmailVerified = true
	err = uc.UserRepo.Save(user)
	if err != nil {
		return fmt.Errorf("Failed to verify email")
	}

	_ = uc.UserTokenRepo.DeleteByToken(token)
	return nil
}

func (uc *AuthUseCase) ResendVerificationEmail(email string) error {
	user, err := uc.UserRepo.FindByEmail(email)
	if err != nil {
		return fmt.Errorf("User not found")
	}

	if user.EmailVerified {
		return fmt.Errorf("Email already verified")
	}

	verificationToken := uniqueEntityId.NewID().String()
	userToken := entity.NewUserToken(user.ID, verificationToken, entity.TokenContextEmailVerification, 24*time.Hour)
	err = uc.UserTokenRepo.Save(userToken)
	if err != nil {
		slog.Error("Failed to save user verification token", "error", err)
		return fmt.Errorf("Failed to generate verification token")
	}

	verifyLink := fmt.Sprintf("/auth/verify-email?token=%s", verificationToken)
	return uc.EmailService.SendVerificationEmail(user.Email, user.Name, verifyLink)
}

func (uc *AuthUseCase) RequestPasswordReset(email string) error {
	user, err := uc.UserRepo.FindByEmail(email)
	if err != nil {
		// Don't leak user existence
		return nil
	}

	resetToken := uniqueEntityId.NewID().String()
	userToken := entity.NewUserToken(user.ID, resetToken, entity.TokenContextPasswordReset, 1*time.Hour)

	err = uc.UserTokenRepo.Save(userToken)
	if err != nil {
		return fmt.Errorf("Failed to generate reset token")
	}

	resetLink := fmt.Sprintf("/auth/reset-password?token=%s", resetToken)
	return uc.EmailService.SendPasswordResetEmail(user.Email, user.Name, resetLink)
}

func (uc *AuthUseCase) ResetPassword(token, newPassword string) error {
	userToken, err := uc.UserTokenRepo.FindByToken(token)
	if err != nil {
		return fmt.Errorf("Invalid or expired token")
	}

	if userToken.Context != entity.TokenContextPasswordReset || userToken.IsExpired() {
		return fmt.Errorf("Invalid or expired token")
	}

	user, err := uc.UserRepo.FindById(userToken.UserID.String())
	if err != nil {
		return fmt.Errorf("User not found")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Failed to reset password")
	}

	user.PasswordHash = string(hashedPassword)
	err = uc.UserRepo.Save(user)
	if err != nil {
		return fmt.Errorf("Failed to update password")
	}

	_ = uc.UserTokenRepo.DeleteByToken(token)
	return nil
}
