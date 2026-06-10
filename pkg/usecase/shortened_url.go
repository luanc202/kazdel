package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"kazdel/pkg/entity"
	"kazdel/pkg/entity/dto"
	"kazdel/pkg/infra/config"
	interfaces "kazdel/pkg/interface"
	"kazdel/pkg/slug"
	"kazdel/pkg/uniqueEntityId"

	"github.com/mssola/user_agent"
	"github.com/oschwald/geoip2-golang"
	"golang.org/x/crypto/bcrypt"
)

type ShortenedUrlUsecase struct {
	repo       interfaces.ShortenedUrlRepository
	visitRepo  interfaces.UrlVisitRepository
	geoipDb    *geoip2.Reader
	logger     slog.Logger
	emailService interfaces.EmailService
}

func NewShortenedUrlUseCase(repo interfaces.ShortenedUrlRepository, visitRepo interfaces.UrlVisitRepository, geoipDb *geoip2.Reader, emailService interfaces.EmailService) *ShortenedUrlUsecase {
	return &ShortenedUrlUsecase{
		repo:       repo,
		visitRepo:  visitRepo,
		geoipDb:    geoipDb,
		logger:     *config.GetLogger("shortened-url-usecase"),
		emailService: emailService,
	}
}

func (su *ShortenedUrlUsecase) Save(shortenedUrlDto dto.ShortenedUrlInsert, userId uniqueEntityId.ID) (string, error) {
	shortSlug := shortenedUrlDto.CustomSlug
	if shortSlug == "" {
		shortSlug = slug.GenerateSlug(6)
	}

	var passwordHash *string
	if shortenedUrlDto.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(shortenedUrlDto.Password), bcrypt.DefaultCost)
		if err != nil {
			return "", fmt.Errorf("failed to hash password: %w", err)
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
		return "", err
	}

	return shortSlug, nil
}

func (su *ShortenedUrlUsecase) FindBySlug(slug string) (*entity.ShortenedUrl, error) {
	return su.repo.FindBySlug(slug)
}

func (su *ShortenedUrlUsecase) ListByUser(userId uniqueEntityId.ID, search string, page, limit int) ([]*entity.ShortenedUrl, int, error) {
	return su.repo.FindByUserIdPaginated(userId, search, page, limit)
}

func (su *ShortenedUrlUsecase) CleanExpiredURLs() error {
	return su.repo.DeleteExpired(time.Now().UTC())
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

func (su *ShortenedUrlUsecase) RecordVisit(ctx context.Context, urlId uint64, req *http.Request) {
	// Extract IP
	ipStr := req.Header.Get("X-Forwarded-For")
	if ipStr == "" {
		ipStr = req.RemoteAddr
	}
	// X-Forwarded-For could be a comma-separated list
	if strings.Contains(ipStr, ",") {
		ipStr = strings.Split(ipStr, ",")[0]
	}
	// Strip port from RemoteAddr
	if strings.Contains(ipStr, ":") {
		host, _, err := net.SplitHostPort(ipStr)
		if err == nil {
			ipStr = host
		}
	}

	userAgentStr := req.UserAgent()
	ua := user_agent.New(userAgentStr)
	browserName, _ := ua.Browser()
	osName := ua.OS()

	referrerStr := req.Referer()

	var countryStr string
	if su.geoipDb != nil && ipStr != "" {
		ip := net.ParseIP(strings.TrimSpace(ipStr))
		if ip != nil {
			record, err := su.geoipDb.Country(ip)
			if err == nil && record.Country.Names != nil {
				countryStr = record.Country.Names["en"]
			}
		}
	}

	var pIp, pReferrer, pUserAgent, pBrowser, pOs, pCountry *string

	if ipStr != "" {
		pIp = &ipStr
	}
	if referrerStr != "" {
		pReferrer = &referrerStr
	}
	if userAgentStr != "" {
		pUserAgent = &userAgentStr
	}
	if browserName != "" {
		pBrowser = &browserName
	}
	if osName != "" {
		pOs = &osName
	}
	if countryStr != "" {
		pCountry = &countryStr
	}

	visit := entity.NewUrlVisit(urlId, pIp, pReferrer, pUserAgent, pBrowser, pOs, pCountry)
	err := su.visitRepo.Save(ctx, visit)
	if err != nil {
		su.logger.Error("Failed to save url visit: " + err.Error())
	}
}

func (su *ShortenedUrlUsecase) GetUrlStats(ctx context.Context, slug string, userId uniqueEntityId.ID) (*dto.UrlStats, error) {
	url, err := su.FindBySlug(slug)
	if err != nil {
		return nil, err
	}
	if url.UserId != userId {
		return nil, fmt.Errorf("unauthorized")
	}

	return su.visitRepo.GetStatsByUrlId(ctx, url.ID)
}

func (su *ShortenedUrlUsecase) ReportUrl(slug, reason, description, reporterEmail string) error {
	// First check if the URL exists
	_, err := su.FindBySlug(slug)
	if err != nil {
		return fmt.Errorf("URL not found")
	}

	// We assume a base URL like domain.com/s/slug
	// But simply returning the slug or formatting it is fine.
	reportedUrl := fmt.Sprintf("/s/%s", slug)

	err = su.emailService.SendReportEmail(reportedUrl, reason, description, reporterEmail)
	if err != nil {
		su.logger.Error("Failed to send report email", "error", err)
		return fmt.Errorf("Failed to submit report")
	}

	return nil
}
