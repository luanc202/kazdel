package interfaces

import "kazdel/pkg/entity"

type UserRepository interface {
	Save(user *entity.User) error
	FindByEmail(email string) (*entity.User, error)
	FindByUsername(username string) (*entity.User, error)
	ExistsByEmail(email string) (bool, error)
	ExistsByUsername(username string) (bool, error)
	FindById(id string) (*entity.User, error)
}
