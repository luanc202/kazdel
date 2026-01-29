package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"url-shortener/m/entity"
	"url-shortener/m/entity/dto"
	"url-shortener/m/internal/uniqueEntityId"
	"url-shortener/m/mocks"
)

func TestShortenedUrlUsecase_Save(t *testing.T) {
	userId, _ := uniqueEntityId.ParseID("1")

	tests := map[string]struct {
		input         dto.ShortenedUrlInsert
		setupMock     func(m *mocks.ShortenedUrlRepository)
		expectedError error
	}{
		"Success": {
			input: dto.ShortenedUrlInsert{
				OriginalUrl: "http://example.com",
				ExpiresAt:   time.Now().Add(30 * time.Minute),
			},
			setupMock: func(m *mocks.ShortenedUrlRepository) {
				m.On("Save", mock.MatchedBy(func(u *entity.ShortenedUrl) bool {
					return u.OriginalUrl == "http://example.com" && u.UserId == userId
				})).Return(nil)
			},
			expectedError: nil,
		},
		"Repository Error": {
			input: dto.ShortenedUrlInsert{
				OriginalUrl: "http://example.com",
				ExpiresAt:   time.Now().Add(30 * time.Minute),
			},
			setupMock: func(m *mocks.ShortenedUrlRepository) {
				m.On("Save", mock.AnythingOfType("*entity.ShortenedUrl")).Return(errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockRepo := new(mocks.ShortenedUrlRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			usecase := NewShortenedUrlUseCase(mockRepo)
			err := usecase.Save(tt.input, userId)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestShortenedUrlUsecase_FindBySlug(t *testing.T) {
	expectedUrl := &entity.ShortenedUrl{
		ShortSlug:   "abc123",
		OriginalUrl: "http://example.com",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	tests := map[string]struct {
		slug           string
		setupMock      func(m *mocks.ShortenedUrlRepository)
		expectedResult *entity.ShortenedUrl
		expectedError  error
	}{
		"Success": {
			slug: "abc123",
			setupMock: func(m *mocks.ShortenedUrlRepository) {
				m.On("FindBySlug", "abc123").Return(expectedUrl, nil)
			},
			expectedResult: expectedUrl,
			expectedError:  nil,
		},
		"Not Found": {
			slug: "unknown",
			setupMock: func(m *mocks.ShortenedUrlRepository) {
				m.On("FindBySlug", "unknown").Return(nil, errors.New("not found"))
			},
			expectedResult: nil,
			expectedError:  errors.New("not found"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockRepo := new(mocks.ShortenedUrlRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			usecase := NewShortenedUrlUseCase(mockRepo)
			result, err := usecase.FindBySlug(tt.slug)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
