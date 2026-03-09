package interfaces

import "kazdel/pkg/entity"

type SessionRepository interface {
	Create(session *entity.Session) error
	FindByToken(token string) (*entity.Session, error)
	DeleteByToken(token string) error
	DeleteExpired() error
}
