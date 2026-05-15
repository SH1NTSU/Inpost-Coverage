package coverage

import (
	"context"
	"sync"
)

type WarmStats struct {
	Provinces int
}

// warmConcurrency caps how many provinces are pre-computed in parallel.
// Each province's Grid call triggers an Overpass terrain fetch, which is the
// slow part of warming. Three concurrent requests cuts total warm wall-time
// roughly threefold without tripping Overpass rate limits — we verified
// 3-parallel works in the competitor importer.
const warmConcurrency = 3

func (s *Service) WarmDefaults(ctx context.Context, minPoints, provinceLimit, cellMeters, recLimit int) (WarmStats, error) {
	if minPoints < 1 {
		minPoints = 1
	}
	if cellMeters <= 0 {
		cellMeters = 800
	}
	if recLimit <= 0 {
		recLimit = 5
	}

	provinces, err := s.Points.ListProvinces(ctx, minPoints)
	if err != nil {
		return WarmStats{}, err
	}
	if provinceLimit > 0 && len(provinces) > provinceLimit {
		provinces = provinces[:provinceLimit]
	}

	stats := WarmStats{Provinces: len(provinces)}
	sem := make(chan struct{}, warmConcurrency)
	var wg sync.WaitGroup
	for _, p := range provinces {
		select {
		case <-ctx.Done():
			wg.Wait()
			return stats, ctx.Err()
		case sem <- struct{}{}:
		}
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			defer func() { <-sem }()
			// Each province's grid + recs warm is independent; errors get
			// logged via the use case's own cache miss path on the next
			// real request, so we swallow them here to let other provinces
			// still warm successfully.
			_, _ = s.Grid(ctx, name, cellMeters)
			_, _ = s.Recommendations(ctx, name, recLimit)
		}(p.Name)
	}
	wg.Wait()
	return stats, nil
}
