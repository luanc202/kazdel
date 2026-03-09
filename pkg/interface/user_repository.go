package interfaces

import "url-shortener/m/pkg/entity"

type UserRepository interface {
	Save(user *entity.User) error
	FindByEmail(email string) (*entity.User, error)
}
