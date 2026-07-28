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
	"kazdel/pkg/mocks"
	"kazdel/pkg/uniqueEntityId"
	"kazdel/pkg/usecase"

	"github.com/stretchr/testify/mock"
)

func TestAuthMiddleware(t *testing.T) {
	mockSessionRepo := new(mocks.SessionRepository)
	mockUserRepo := new(mocks.UserRepository)

	validUserId := uniqueEntityId.NewID()
	validToken := "valid-session-token"
	expiredToken := "expired-session-token"

	mockSessionRepo.On("FindByToken", validToken).Return(entity.NewSession(validUserId, validToken, time.Now().Add(1*time.Hour)), nil)

	expiredSession := entity.NewSession(validUserId, expiredToken, time.Now().Add(-1*time.Hour))
	mockSessionRepo.On("FindByToken", expiredToken).Return(expiredSession, nil)
	mockSessionRepo.On("DeleteByToken", expiredToken).Return(nil)

	mockSessionRepo.On("FindByToken", mock.Anything).Return(nil, errors.New("Session not found"))

	mockUserTokenRepo := new(mocks.UserTokenRepository)
	mockEmailService := new(mocks.EmailService)

	authUseCase := usecase.NewAuthUseCase(mockUserRepo, mockSessionRepo, mockUserTokenRepo, mockEmailService)

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
		name         string
		cookieValue  string
		expectStatus int
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
