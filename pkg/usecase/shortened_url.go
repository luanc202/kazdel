package usecase

import (
	"errors"
	"fmt"
	"log/slog"
	"kazdel/pkg/entity"
	"kazdel/pkg/entity/dto"
	"kazdel/pkg/infra/config"
	interfaces "kazdel/pkg/interface"
	"kazdel/pkg/slug"
	"kazdel/pkg/uniqueEntityId"

	"golang.org/x/crypto/bcrypt"
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
	shortSlug := shortenedUrlDto.CustomSlug
	if shortSlug == "" {
		shortSlug = slug.GenerateSlug(6)
	}

	var passwordHash *string
	if shortenedUrlDto.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(shortenedUrlDto.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		hashStr := string(hash)
		passwordHash = &hashStr
	}

	var description *string
	if shortenedUrlDto.Description != "" {
		descStr := shortenedUrlDto.Description
		description = &descStr
	}

	expiresAt := shortenedUrlDto.ParsedExpiresAt

	shortenedUrl := entity.NewShortenedUrl(shortSlug, shortenedUrlDto.OriginalUrl, expiresAt, userId, description, passwordHash)

	err := su.repo.Save(shortenedUrl)

	if err != nil {
		if !errors.Is(err, entity.ErrCustomSlugAlreadyExists) {
			su.logger.Error(fmt.Errorf("Error saving shortened url: %w", err).Error())
		}
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

func (su *ShortenedUrlUsecase) Update(slug string, updateDto dto.ShortenedUrlUpdate, userId uniqueEntityId.ID) error {
	existingUrl, err := su.repo.FindBySlug(slug)
	if err != nil {
		return err
	}
	if existingUrl.UserId != userId {
		return fmt.Errorf("unauthorized to edit this url")
	}

	var passwordHash *string
	if updateDto.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(updateDto.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		hashStr := string(hash)
		passwordHash = &hashStr
	} else {
		// if empty, we might want to keep the existing one or clear it?
		// Let's assume empty means clear it, or maybe a special value.
		// Actually, if they want to clear it, they might leave it empty. Let's assume empty clears it for simplicity.
		passwordHash = nil
	}

	var description *string
	if updateDto.Description != "" {
		descStr := updateDto.Description
		description = &descStr
	}

	existingUrl.LongUrl = updateDto.OriginalUrl
	existingUrl.Description = description
	existingUrl.PasswordHash = passwordHash
	existingUrl.ExpiresAt = updateDto.ParsedExpiresAt

	err = su.repo.Update(existingUrl)
	if err != nil {
		su.logger.Error(fmt.Errorf("Error updating shortened url: %w", err).Error())
		return err
	}

	return nil
}
