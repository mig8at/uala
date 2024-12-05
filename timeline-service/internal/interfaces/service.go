package interfaces

import (
	"context"
	"timeline-service/internal/domain/models"
)

type Service interface {
	Paginate(ctx context.Context, id string, limit, offset int) ([]*models.Tweet, error)
	SyncData(ctx context.Context) error
}
