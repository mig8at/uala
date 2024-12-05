package application

import (
	"context"
	"timeline-service/internal/domain/models"
	"timeline-service/internal/interfaces"
)

type timelineService struct {
	repo interfaces.Repository
}

func NewService(repo interfaces.Repository) interfaces.Service {
	return &timelineService{
		repo: repo,
	}
}

func (s *timelineService) Paginate(ctx context.Context, id string, page, size int) ([]*models.Timeline, error) {
	return s.repo.Paginate(ctx, id, page, size)
}
