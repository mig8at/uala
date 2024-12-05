package interfaces

import (
	"context"
	"timeline-service/internal/domain/models"
)

type Repository interface {
	Paginate(ctx context.Context, id string, page, size int) ([]*models.Timeline, error)
}
