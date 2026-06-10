package mocks

import (
	"kazdel/pkg/entity"
	"github.com/stretchr/testify/mock"
)

type UserTokenRepository struct {
	mock.Mock
}

func (_m *UserTokenRepository) Save(token *entity.UserToken) error {
	ret := _m.Called(token)
	return ret.Error(0)
}

func (_m *UserTokenRepository) FindByToken(token string) (*entity.UserToken, error) {
	ret := _m.Called(token)
	var r0 *entity.UserToken
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*entity.UserToken)
	}
	return r0, ret.Error(1)
}

func (_m *UserTokenRepository) DeleteByToken(token string) error {
	ret := _m.Called(token)
	return ret.Error(0)
}

func (_m *UserTokenRepository) DeleteByUserIdAndContext(userId string, context entity.TokenContext) error {
	ret := _m.Called(userId, context)
	return ret.Error(0)
}

func NewUserTokenRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *UserTokenRepository {
	m := &UserTokenRepository{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}
