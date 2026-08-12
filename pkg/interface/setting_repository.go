package interfaces

import "kazdel/pkg/entity"

type SettingRepository interface {
	Save(setting *entity.Setting) error
	FindByKey(key string) (*entity.Setting, error)
	FindAll() ([]*entity.Setting, error)
}
