package scraping

import (
	"context"
	"fmt"
	"time"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
)

type Service struct {
	Client    domain.InpostClient
	Points    domain.PointRepository
	Snapshots domain.SnapshotRepository

	Country  string
	City     string
	PageSize int
}

func (s *Service) RunOnce(ctx context.Context) (Stats, error) {
	stats := Stats{StartedAt: time.Now()}
	err := s.Client.ListByCity(ctx, s.Country, s.City, s.PageSize,
		func(page []domain.Point, snapshots []domain.AvailabilitySnapshot) error {
			for i := range page {
				if err := s.Points.Upsert(ctx, &page[i]); err != nil {
					return fmt.Errorf("upsert point %s: %w", page[i].InpostID, err)
				}
				snapshots[i].PointID = page[i].ID
			}
			if err := s.Snapshots.InsertBatch(ctx, snapshots); err != nil {
				return fmt.Errorf("insert snapshots: %w", err)
			}
			stats.Points += len(page)
			stats.Snapshots += len(snapshots)
			return nil
		})
	stats.FinishedAt = time.Now()
	return stats, err
}

type Stats struct {
	StartedAt  time.Time
	FinishedAt time.Time
	Points     int
	Snapshots  int
}
