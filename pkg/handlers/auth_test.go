package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kazdel/pkg/constants"
	"kazdel/pkg/entity"
	"kazdel/pkg/usecase"
)

// Minimal mock repos for Auth handler tests
type MockUserRepository struct{}

func (m *MockUserRepository) Save(user *entity.User) error                         { return nil }
func (m *MockUserRepository) FindByEmail(email string) (*entity.User, error)       { return nil, nil }
func (m *MockUserRepository) FindByUsername(username string) (*entity.User, error) { return nil, nil }
func (m *MockUserRepository) ExistsByEmail(email string) (bool, error)             { return false, nil }
func (m *MockUserRepository) ExistsByUsername(username string) (bool, error)       { return false, nil }

type MockSessionRepository struct {
	sessions map[string]*entity.Session
}

func (m *MockSessionRepository) Create(session *entity.Session) error { return nil }
func (m *MockSessionRepository) FindByToken(token string) (*entity.Session, error) {
	return nil, errors.New("Not implemented")
}
func (m *MockSessionRepository) DeleteByToken(token string) error {
	delete(m.sessions, token)
	return nil
}
func (m *MockSessionRepository) DeleteExpired() error { return nil }

func TestAuth_Logout(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	mockSessionRepo := &MockSessionRepository{
		sessions: make(map[string]*entity.Session),
	}

	authUseCase := usecase.NewAuthUseCase(mockUserRepo, mockSessionRepo)
	authHandler := &Auth{authUseCase: authUseCase}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)
	
	validToken := "my-session-token"
	mockSessionRepo.sessions[validToken] = &entity.Session{Token: validToken}

	req.AddCookie(&http.Cookie{
		Name:  constants.SessionCookieName,
		Value: validToken,
	})

	rr := httptest.NewRecorder()
	authHandler.Logout(rr, req)

	// It should respond with 200 OK because HTMX receives 200 and processes HX-Redirect
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %v", rr.Code)
	}

	// It should set HX-Redirect header
	if hxRedirect := rr.Header().Get("HX-Redirect"); hxRedirect != "/login" {
		t.Errorf("Expected HX-Redirect to /login, got %v", hxRedirect)
	}

	// It should set a cookie that expires in the past
	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie set, got %v", len(cookies))
	}
	
	if cookies[0].Name != constants.SessionCookieName {
		t.Errorf("Expected cookie name %s, got %s", constants.SessionCookieName, cookies[0].Name)
	}
	
	if cookies[0].Value != "" {
		t.Errorf("Expected empty cookie value, got %v", cookies[0].Value)
	}

	if cookies[0].Expires.After(time.Now()) {
		t.Errorf("Expected cookie expiration to be in the past, got %v", cookies[0].Expires)
	}

	// Ensure the session was truly deleted from the repo
	if _, stillExists := mockSessionRepo.sessions[validToken]; stillExists {
		t.Errorf("Expected session to be deleted from repo")
	}
}
