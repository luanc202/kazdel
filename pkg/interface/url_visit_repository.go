package interfaces

import (
	"context"

	"kazdel/pkg/entity"
	"kazdel/pkg/entity/dto"
)

type UrlVisitRepository interface {
	Save(ctx context.Context, visit *entity.UrlVisit) error
	GetStatsByUrlId(ctx context.Context, urlId uint64) (*dto.UrlStats, error)
}
