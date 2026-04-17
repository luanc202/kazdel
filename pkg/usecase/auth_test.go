package usecase

import (
	"errors"
	"testing"
	"time"

	"kazdel/pkg/entity"
	"kazdel/pkg/mocks"
	"kazdel/pkg/uniqueEntityId"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthUseCase_Signup(t *testing.T) {
	tests := map[string]struct {
		name          string
		username      string
		email         string
		password      string
		setupMock     func(u *mocks.UserRepository, s *mocks.SessionRepository)
		expectedError string
	}{
		"Email already registered": {
			email:    "test@example.com",
			password: "password123",
			setupMock: func(u *mocks.UserRepository, s *mocks.SessionRepository) {
				u.On("ExistsByEmail", "test@example.com").Return(true, nil)
			},
			expectedError: "Email already registered",
		},
		"Username already registered": {
			username: "testuser",
			email:    "test@example.com",
			password: "password123",
			setupMock: func(u *mocks.UserRepository, s *mocks.SessionRepository) {
				u.On("ExistsByEmail", "test@example.com").Return(false, nil)
				u.On("ExistsByUsername", "testuser").Return(true, nil)
			},
			expectedError: "Username already registered",
		},
		"Success": {
			name:     "Test User",
			username: "testuser",
			email:    "test@example.com",
			password: "password123",
			setupMock: func(u *mocks.UserRepository, s *mocks.SessionRepository) {
				u.On("ExistsByEmail", "test@example.com").Return(false, nil)
				u.On("ExistsByUsername", "testuser").Return(false, nil)
				u.On("Save", mock.AnythingOfType("*entity.User")).Return(nil)
				s.On("Create", mock.AnythingOfType("*entity.Session")).Return(nil)
			},
			expectedError: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockUserRepo := new(mocks.UserRepository)
			mockSessionRepo := new(mocks.SessionRepository)

			tt.setupMock(mockUserRepo, mockSessionRepo)

			uc := NewAuthUseCase(mockUserRepo, mockSessionRepo)
			token, err := uc.Signup(tt.name, tt.username, tt.email, tt.password)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}

			mockUserRepo.AssertExpectations(t)
			mockSessionRepo.AssertExpectations(t)
		})
	}
}

func TestAuthUseCase_Login(t *testing.T) {
	validPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)

	validUser := &entity.User{
		ID:           uniqueEntityId.NewID(),
		Username:     "testuser",
		PasswordHash: string(hashedPassword),
	}

	tests := map[string]struct {
		username      string
		password      string
		setupMock     func(u *mocks.UserRepository, s *mocks.SessionRepository)
		expectedError string
	}{
		"Invalid Username": {
			username: "wronguser",
			password: validPassword,
			setupMock: func(u *mocks.UserRepository, s *mocks.SessionRepository) {
				u.On("FindByUsername", "wronguser").Return(nil, errors.New("not found"))
			},
			expectedError: "Invalid credentials",
		},
		"Invalid Password": {
			username: "testuser",
			password: "wrongpassword",
			setupMock: func(u *mocks.UserRepository, s *mocks.SessionRepository) {
				u.On("FindByUsername", "testuser").Return(validUser, nil)
			},
			expectedError: "Invalid credentials",
		},
		"Success": {
			username: "testuser",
			password: validPassword,
			setupMock: func(u *mocks.UserRepository, s *mocks.SessionRepository) {
				u.On("FindByUsername", "testuser").Return(validUser, nil)
				s.On("Create", mock.AnythingOfType("*entity.Session")).Return(nil)
			},
			expectedError: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockUserRepo := new(mocks.UserRepository)
			mockSessionRepo := new(mocks.SessionRepository)

			tt.setupMock(mockUserRepo, mockSessionRepo)

			uc := NewAuthUseCase(mockUserRepo, mockSessionRepo)
			token, err := uc.Login(tt.username, tt.password)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}

			mockUserRepo.AssertExpectations(t)
			mockSessionRepo.AssertExpectations(t)
		})
	}
}

func TestAuthUseCase_Logout(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockSessionRepo := new(mocks.SessionRepository)

	mockSessionRepo.On("DeleteByToken", "validtoken").Return(nil)

	uc := NewAuthUseCase(mockUserRepo, mockSessionRepo)
	err := uc.Logout("validtoken")

	assert.NoError(t, err)
	mockSessionRepo.AssertExpectations(t)
}

func TestAuthUseCase_ValidateSession(t *testing.T) {
	validToken := "valid-token"
	expiredToken := "expired-token"
	userId := uniqueEntityId.NewID()

	tests := map[string]struct {
		token         string
		setupMock     func(s *mocks.SessionRepository)
		expectedError string
	}{
		"Session Not Found (Error)": {
			token: "wrong-token",
			setupMock: func(s *mocks.SessionRepository) {
				s.On("FindByToken", "wrong-token").Return(nil, errors.New("db error"))
			},
			expectedError: "db error",
		},
		"Session Not Found (Nil)": {
			token: "wrong-token-nil",
			setupMock: func(s *mocks.SessionRepository) {
				s.On("FindByToken", "wrong-token-nil").Return(nil, nil)
			},
			expectedError: "Session not found",
		},
		"Session Expired": {
			token: expiredToken,
			setupMock: func(s *mocks.SessionRepository) {
				session := entity.NewSession(userId, expiredToken, time.Now().Add(-1*time.Hour))
				s.On("FindByToken", expiredToken).Return(session, nil)
				s.On("DeleteByToken", expiredToken).Return(nil)
			},
			expectedError: "Session expired",
		},
		"Success": {
			token: validToken,
			setupMock: func(s *mocks.SessionRepository) {
				session := entity.NewSession(userId, validToken, time.Now().Add(1*time.Hour))
				s.On("FindByToken", validToken).Return(session, nil)
			},
			expectedError: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockUserRepo := new(mocks.UserRepository)
			mockSessionRepo := new(mocks.SessionRepository)

			tt.setupMock(mockSessionRepo)

			uc := NewAuthUseCase(mockUserRepo, mockSessionRepo)
			returnedUserId, err := uc.ValidateSession(tt.token)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				assert.Empty(t, returnedUserId)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, userId.String(), returnedUserId)
			}

			mockSessionRepo.AssertExpectations(t)
		})
	}
}
