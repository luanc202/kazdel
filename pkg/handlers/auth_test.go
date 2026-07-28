package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"kazdel/pkg/constants"
	"kazdel/pkg/entity"
	"kazdel/pkg/infra/config"
	"kazdel/pkg/mocks"
	"kazdel/pkg/usecase"

	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuth_Logout(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockSessionRepo := new(mocks.SessionRepository)
	mockUserTokenRepo := new(mocks.UserTokenRepository)
	mockEmailService := new(mocks.EmailService)

	authUseCase := usecase.NewAuthUseCase(mockUserRepo, mockSessionRepo, mockUserTokenRepo, mockEmailService)
	authHandler := &Auth{authUseCase: authUseCase}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)

	validToken := "my-session-token"

	mockSessionRepo.On("DeleteByToken", validToken).Return(nil)

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

	mockSessionRepo.AssertExpectations(t)
}

func TestAuth_SignupSubmit(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockSessionRepo := new(mocks.SessionRepository)
	mockUserTokenRepo := new(mocks.UserTokenRepository)
	mockEmailService := new(mocks.EmailService)

	authUseCase := usecase.NewAuthUseCase(mockUserRepo, mockSessionRepo, mockUserTokenRepo, mockEmailService)
	authHandler := &Auth{authUseCase: authUseCase}

	// Setup mock expectations
	mockUserRepo.On("ExistsByEmail", "test@example.com").Return(false, nil)
	mockUserRepo.On("ExistsByUsername", "testuser").Return(false, nil)
	mockUserRepo.On("Save", mock.AnythingOfType("*entity.User")).Return(nil)
	mockSessionRepo.On("Create", mock.AnythingOfType("*entity.Session")).Return(nil)
	mockUserTokenRepo.On("Save", mock.AnythingOfType("*entity.UserToken")).Return(nil)
	mockEmailService.On("SendVerificationEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	formData := "name=Test+User&username=testuser&email=test%40example.com&password=Password1%21"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	authHandler.SignupSubmit(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %v", rr.Code)
	}

	if hxRedirect := rr.Header().Get("HX-Redirect"); hxRedirect != "/dashboard" {
		t.Errorf("Expected HX-Redirect to /dashboard, got %v", hxRedirect)
	}

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie set, got %v", len(cookies))
	}

	if cookies[0].Name != constants.SessionCookieName {
		t.Errorf("Expected cookie name %s, got %s", constants.SessionCookieName, cookies[0].Name)
	}

	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestAuth_LoginSubmit(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockSessionRepo := new(mocks.SessionRepository)
	mockUserTokenRepo := new(mocks.UserTokenRepository)
	mockEmailService := new(mocks.EmailService)

	authUseCase := usecase.NewAuthUseCase(mockUserRepo, mockSessionRepo, mockUserTokenRepo, mockEmailService)
	authHandler := &Auth{authUseCase: authUseCase}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Password1!"), bcrypt.DefaultCost)
	testUser := entity.NewUser("Test User", "testuser", entity.RoleUser, "test@example.com", string(hashedPassword))

	mockUserRepo.On("FindByUsername", "testuser").Return(testUser, nil)
	mockSessionRepo.On("Create", mock.AnythingOfType("*entity.Session")).Return(nil)

	formData := "username=testuser&password=Password1%21"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	authHandler.LoginSubmit(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %v", rr.Code)
	}

	if hxRedirect := rr.Header().Get("HX-Redirect"); hxRedirect != "/dashboard" {
		t.Errorf("Expected HX-Redirect to /dashboard, got %v", hxRedirect)
	}

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie set, got %v", len(cookies))
	}

	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestAuth_LoginSubmit_EmailNotVerified(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockSessionRepo := new(mocks.SessionRepository)
	mockUserTokenRepo := new(mocks.UserTokenRepository)
	mockEmailService := new(mocks.EmailService)

	authUseCase := usecase.NewAuthUseCase(mockUserRepo, mockSessionRepo, mockUserTokenRepo, mockEmailService)
	authHandler := &Auth{authUseCase: authUseCase}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Password1!"), bcrypt.DefaultCost)
	testUser := entity.NewUser("Test User", "testuser", entity.RoleUser, "test@example.com", string(hashedPassword))
	testUser.EmailVerified = false

	oldConfig := config.GetEnvConfig()
	config.SetEnvConfigForTest(&config.EnvConfig{MAIL_ENABLED: true})
	defer config.SetEnvConfigForTest(oldConfig)

	mockUserRepo.On("FindByUsername", "testuser").Return(testUser, nil)

	formData := "username=testuser&password=Password1%21"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	authHandler.LoginSubmit(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %v", rr.Code)
	}

	if hxRedirect := rr.Header().Get("HX-Redirect"); hxRedirect != "" {
		t.Errorf("Expected empty HX-Redirect, got %v", hxRedirect)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Your email is not verified.") {
		t.Errorf("Expected response body to contain 'Your email is not verified.', got %v", body)
	}

	mockUserRepo.AssertExpectations(t)
}

func TestAuth_VerifyEmailPage(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockSessionRepo := new(mocks.SessionRepository)
	mockUserTokenRepo := new(mocks.UserTokenRepository)
	mockEmailService := new(mocks.EmailService)

	authUseCase := usecase.NewAuthUseCase(mockUserRepo, mockSessionRepo, mockUserTokenRepo, mockEmailService)
	authHandler := &Auth{authUseCase: authUseCase}

	req := httptest.NewRequest(http.MethodGet, "/verify-email?email=test%40example.com", nil)
	rr := httptest.NewRecorder()
	authHandler.VerifyEmailPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %v", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "VERIFY EMAIL") {
		t.Errorf("Expected response body to contain 'VERIFY EMAIL', got %v", body)
	}
	if !strings.Contains(body, "test@example.com") {
		t.Errorf("Expected response body to contain 'test@example.com', got %v", body)
	}
}

func TestAuth_ResendVerificationSubmit(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockSessionRepo := new(mocks.SessionRepository)
	mockUserTokenRepo := new(mocks.UserTokenRepository)
	mockEmailService := new(mocks.EmailService)

	authUseCase := usecase.NewAuthUseCase(mockUserRepo, mockSessionRepo, mockUserTokenRepo, mockEmailService)
	authHandler := &Auth{authUseCase: authUseCase}

	testUser := entity.NewUser("Test User", "testuser", entity.RoleUser, "test@example.com", "hash")
	testUser.EmailVerified = false

	mockUserRepo.On("FindByEmail", "test@example.com").Return(testUser, nil)
	mockUserTokenRepo.On("Save", mock.AnythingOfType("*entity.UserToken")).Return(nil)
	mockEmailService.On("SendVerificationEmail", "test@example.com", "Test User", mock.AnythingOfType("string")).Return(nil)

	formData := "email=test%40example.com"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/resend-verification", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	authHandler.ResendVerificationSubmit(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %v", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "alert-success") {
		t.Errorf("Expected success message alert in body, got %v", body)
	}

	mockUserRepo.AssertExpectations(t)
	mockUserTokenRepo.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}
