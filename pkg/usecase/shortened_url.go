package usecase

import (
	"fmt"
	"log/slog"
	"url-shortener/m/pkg/entity"
	"url-shortener/m/pkg/entity/dto"
	"url-shortener/m/pkg/infra/config"
	interfaces "url-shortener/m/pkg/interface"
	"url-shortener/m/pkg/slug"
	"url-shortener/m/pkg/uniqueEntityId"
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
	shortSlug := slug.GenerateSlug(6)

	expiresAt := shortenedUrlDto.ExpiresAt

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

func (su *ShortenedUrlUsecase) ListByUser(userId uniqueEntityId.ID) ([]*entity.ShortenedUrl, error) {
	return su.repo.FindByUserId(userId)
}

func (su *ShortenedUrlUsecase) Delete(id uint64, userId uniqueEntityId.ID) error {
	return su.repo.Delete(id, userId)
}
