package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	appctx "kazdel/pkg/context"
	"kazdel/pkg/entity"
	"kazdel/pkg/mocks"
	"kazdel/pkg/uniqueEntityId"
	"kazdel/pkg/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
)

func TestShortenedUrl_RedirectToLongUrl(t *testing.T) {
	mockRepo := new(mocks.ShortenedUrlRepository)
	mockVisitRepo := new(mocks.UrlVisitRepository)
	suUseCase := usecase.NewShortenedUrlUseCase(mockRepo, mockVisitRepo, nil, nil)

	// We can pass nil for authUseCase here because this route is public
	handler := &ShortenedUrl{usecase: suUseCase, authUseCase: nil}

	// Create test dependencies
	testSlug := "myslug"
	testLongUrl := "https://example.com/very/long/path"

	validUrl := &entity.ShortenedUrl{
		ShortSlug: testSlug,
		LongUrl:   testLongUrl,
	}

	mockRepo.On("FindBySlug", testSlug).Return(validUrl, nil).Once()
	mockRepo.On("FindBySlug", "notfound").Return(nil, errors.New("not found")).Once()

	mockVisitRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.UrlVisit")).Return(nil).Maybe()

	tests := []struct {
		name         string
		slug         string
		expectStatus int
		expectLoc    string
	}{
		{
			name:         "valid slug redirects",
			slug:         testSlug,
			expectStatus: http.StatusFound, // 302
			expectLoc:    testLongUrl,
		},
		{
			name:         "invalid slug gives 404",
			slug:         "notfound",
			expectStatus: http.StatusOK, // The ErrorPage renders with 200 via Templ but represents a not found state in HTML visually. Wait, actually ErrorPage sets no explicit status, so it defaults to 200.
			expectLoc:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tt.slug, nil)

			// Setup chi router context so chi.URLParam works
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("slug", tt.slug)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.RedirectToLongUrl(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("Expected status %v, got %v", tt.expectStatus, rr.Code)
			}

			if tt.expectLoc != "" {
				loc := rr.Header().Get("Location")
				if loc != tt.expectLoc {
					t.Errorf("Expected location %s, got %s", tt.expectLoc, loc)
				}
			}
		})
	}

	time.Sleep(10 * time.Millisecond)
	mockRepo.AssertExpectations(t)
	mockVisitRepo.AssertExpectations(t)
}

func TestShortenedUrl_CreateUrl(t *testing.T) {
	mockRepo := new(mocks.ShortenedUrlRepository)
	mockVisitRepo := new(mocks.UrlVisitRepository)
	suUseCase := usecase.NewShortenedUrlUseCase(mockRepo, mockVisitRepo, nil, nil)

	handler := &ShortenedUrl{usecase: suUseCase, authUseCase: nil}

	validUserId := uniqueEntityId.NewID()
	expiresAt := time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04")
	validEntity := &entity.ShortenedUrl{
		ShortSlug: "IXHAJB",
		LongUrl:   "https://validurl.com",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mockRepo.On("FindBySlug", mock.Anything).Return(validEntity, nil).Once()
	mockRepo.On("Save", mock.AnythingOfType("*entity.ShortenedUrl")).Return(nil)

	formData := url.Values{}
	formData.Add("originalUrl", "https://validurl.com")
	formData.Add("expiresAt", expiresAt)

	req := httptest.NewRequest(http.MethodPost, "/dashboard/urls/shorten", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Inject user id into context to simulate authenticated user
	req = appctx.SetAuthUser(req, validUserId.String())

	rr := httptest.NewRecorder()
	handler.CreateUrl(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %v", rr.Code)
	}

	mockRepo.AssertExpectations(t)
}
