package interfaces

import (
	"context"
	"timeline-service/internal/domain/models"
)

type Service interface {
	Paginate(ctx context.Context, id string, page, size int) ([]*models.Timeline, error)
}
