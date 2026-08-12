package mocks

import (
	"kazdel/pkg/entity"

	"github.com/stretchr/testify/mock"
)

type MockSettingRepository struct {
	mock.Mock
}

func (m *MockSettingRepository) Save(setting *entity.Setting) error {
	args := m.Called(setting)
	return args.Error(0)
}

func (m *MockSettingRepository) FindByKey(key string) (*entity.Setting, error) {
	args := m.Called(key)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.Setting), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSettingRepository) FindAll() ([]*entity.Setting, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*entity.Setting), args.Error(1)
	}
	return nil, args.Error(1)
}
