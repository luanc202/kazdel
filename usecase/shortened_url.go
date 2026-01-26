package usecase

import (
	"fmt"
	"log/slog"
	"time"
	"url-shortener/m/entity"
	"url-shortener/m/entity/dto"
	"url-shortener/m/infra/config"
	interfaces "url-shortener/m/interface"
	"url-shortener/m/internal"
	"url-shortener/m/internal/uniqueEntityId"
)

type ShortenedUrlUsecase struct {
	repo   interfaces.ShortenedUrlRepository
	logger slog.Logger
}

func NewShortenedUrlUseCase(repo interfaces.ShortenedUrlRepository) *ShortenedUrlUsecase {
	return &ShortenedUrlUsecase{
		repo:   repo,
		logger: *config.GetLogger("shortened-url-usecase"),
	}
}

func (su *ShortenedUrlUsecase) Save(shortenedUrlDto dto.ShortenedUrlInsert, userId uniqueEntityId.ID) error {
	shortSlug := internal.GenerateSlug(6)

	expiresAt := time.Now().Add(time.Duration(shortenedUrlDto.ExpiresAt) * time.Hour * 24 * 30)

	shortenedUrl := entity.NewShortenedUrl(shortSlug, shortenedUrlDto.OriginalUrl, expiresAt, userId)

	err := su.repo.Save(shortenedUrl)

	if err != nil {
		su.logger.Error(fmt.Errorf("Error saving shortened url: %w", err).Error())
		return err
	}

	return nil
}

func (su *ShortenedUrlUsecase) FindBySlug(slug string) (*entity.ShortenedUrl, error) {
	return su.repo.FindBySlug(slug)
}
