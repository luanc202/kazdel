package usecase

import (
	"kazdel/pkg/entity"
	interfaces "kazdel/pkg/interface"

	"golang.org/x/crypto/bcrypt"
)

type AdminUseCase struct {
	userRepo         interfaces.UserRepository
	settingRepo      interfaces.SettingRepository
	shortenedUrlRepo interfaces.ShortenedUrlRepository
}

func NewAdminUseCase(userRepo interfaces.UserRepository, settingRepo interfaces.SettingRepository, shortenedUrlRepo interfaces.ShortenedUrlRepository) *AdminUseCase {
	return &AdminUseCase{
		userRepo:         userRepo,
		settingRepo:      settingRepo,
		shortenedUrlRepo: shortenedUrlRepo,
	}
}

// User Management
func (uc *AdminUseCase) GetUsers(search string, page, limit int) ([]*entity.User, int, error) {
	return uc.userRepo.FindAllPaginated(search, page, limit)
}

func (uc *AdminUseCase) UpdateUserStatus(userId string, isActive bool) error {
	user, err := uc.userRepo.FindById(userId)
	if err != nil {
		return err
	}
	user.IsActive = isActive
	return uc.userRepo.Save(user)
}

func (uc *AdminUseCase) UpdateUserPassword(userId string, newPassword string) error {
	user, err := uc.userRepo.FindById(userId)
	if err != nil {
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)
	return uc.userRepo.Save(user)
}

// Settings Management
func (uc *AdminUseCase) GetSettings() ([]*entity.Setting, error) {
	return uc.settingRepo.FindAll()
}

func (uc *AdminUseCase) UpdateSetting(key, value string) error {
	setting, err := uc.settingRepo.FindByKey(key)
	if err != nil {
		setting = &entity.Setting{
			Key:   key,
			Value: value,
		}
	} else {
		setting.Value = value
	}

	return uc.settingRepo.Save(setting)
}

// URL Management
func (uc *AdminUseCase) GetURLs(search string, page, limit int) ([]*entity.ShortenedUrl, int, error) {
	return uc.shortenedUrlRepo.FindAllPaginated(search, page, limit)
}

func (uc *AdminUseCase) DeleteURL(slug string) error {
	url, err := uc.shortenedUrlRepo.FindBySlug(slug)
	if err != nil {
		return err
	}
	return uc.shortenedUrlRepo.Delete(url.ID, url.UserId)
}
