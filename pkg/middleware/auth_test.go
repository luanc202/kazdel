package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kazdel/pkg/constants"
	appctx "kazdel/pkg/context"
	"kazdel/pkg/entity"
	"kazdel/pkg/uniqueEntityId"
	"kazdel/pkg/usecase"
)

// MockUserRepository implements interfaces.UserRepository
type MockUserRepository struct{}

func (m *MockUserRepository) Save(user *entity.User) error                                  { return nil }
func (m *MockUserRepository) FindByEmail(email string) (*entity.User, error)                { return nil, nil }
func (m *MockUserRepository) FindByUsername(username string) (*entity.User, error)          { return nil, nil }
func (m *MockUserRepository) ExistsByEmail(email string) (bool, error)                      { return false, nil }
func (m *MockUserRepository) ExistsByUsername(username string) (bool, error)                { return false, nil }

// MockSessionRepository implements interfaces.SessionRepository
type MockSessionRepository struct {
	sessions map[string]*entity.Session
}

func (m *MockSessionRepository) Create(session *entity.Session) error {
	m.sessions[session.Token] = session
	return nil
}

func (m *MockSessionRepository) FindByToken(token string) (*entity.Session, error) {
	session, exists := m.sessions[token]
	if !exists {
		return nil, errors.New("Session not found")
	}
	return session, nil
}

func (m *MockSessionRepository) DeleteByToken(token string) error {
	delete(m.sessions, token)
	return nil
}

func (m *MockSessionRepository) DeleteExpired() error { return nil }

func TestAuthMiddleware(t *testing.T) {
	mockSessionRepo := &MockSessionRepository{
		sessions: make(map[string]*entity.Session),
	}
	mockUserRepo := &MockUserRepository{}
	
	// Create a valid session in the mock DB
	validUserId := uniqueEntityId.NewID()
	validToken := "valid-session-token"
	mockSessionRepo.sessions[validToken] = entity.NewSession(validUserId, validToken, time.Now().Add(1*time.Hour))

	expiredToken := "expired-session-token"
	mockSessionRepo.sessions[expiredToken] = entity.NewSession(validUserId, expiredToken, time.Now().Add(-1*time.Hour))

	authUseCase := usecase.NewAuthUseCase(mockUserRepo, mockSessionRepo)

	// Create our middleware
	mw := AuthMiddleware(authUseCase)

	// Create a dummy next handler that writes 200 OK and asserts the context
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context was populated correctly
		uid, ok := appctx.GetAuthUser(r)
		if !ok || uid != validUserId.String() {
			t.Errorf("Expected user id %s in context, got %v", validUserId.String(), uid)
			http.Error(w, "invalid context", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	})

	handlerToTest := mw(nextHandler)

	tests := []struct {
		name           string
		cookieValue    string
		expectStatus   int
	}{
		{
			name:         "valid token",
			cookieValue:  validToken,
			expectStatus: http.StatusOK,
		},
		{
			name:         "invalid token",
			cookieValue:  "wrong-token",
			expectStatus: http.StatusSeeOther, // Expect redirect to login
		},
		{
			name:         "expired token",
			cookieValue:  expiredToken,
			expectStatus: http.StatusSeeOther,
		},
		{
			name:         "missing token",
			cookieValue:  "",
			expectStatus: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
			
			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{
					Name:  constants.SessionCookieName,
					Value: tt.cookieValue,
				})
			}

			rr := httptest.NewRecorder()

			handlerToTest.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("expected status %d, got %d", tt.expectStatus, rr.Code)
			}
			
			if tt.expectStatus == http.StatusSeeOther {
				loc := rr.Header().Get("Location")
				if loc != "/login" {
					t.Errorf("expected redirect to /login, got %s", loc)
				}
			}
		})
	}
}
