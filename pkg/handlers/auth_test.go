package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kazdel/pkg/constants"
	"kazdel/pkg/mocks"
	"kazdel/pkg/usecase"
)

func TestAuth_Logout(t *testing.T) {
	mockUserRepo := new(mocks.UserRepository)
	mockSessionRepo := new(mocks.SessionRepository)

	authUseCase := usecase.NewAuthUseCase(mockUserRepo, mockSessionRepo)
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
