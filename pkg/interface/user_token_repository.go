package interfaces

import "kazdel/pkg/entity"

type UserTokenRepository interface {
	Save(token *entity.UserToken) error
	FindByToken(token string) (*entity.UserToken, error)
	DeleteByToken(token string) error
	DeleteByUserIdAndContext(userId string, context entity.TokenContext) error
}
