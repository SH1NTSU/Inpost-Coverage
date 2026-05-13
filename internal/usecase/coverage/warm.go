package coverage

import "context"

type WarmStats struct {
	Cities int
}

func (s *Service) WarmDefaults(ctx context.Context, minPoints, cityLimit, cellMeters, recLimit int) (WarmStats, error) {
	if minPoints < 1 {
		minPoints = 1
	}
	if cellMeters <= 0 {
		cellMeters = 400
	}
	if recLimit <= 0 {
		recLimit = 5
	}

	cities, err := s.Points.ListCities(ctx, minPoints)
	if err != nil {
		return WarmStats{}, err
	}
	if cityLimit > 0 && len(cities) > cityLimit {
		cities = cities[:cityLimit]
	}

	stats := WarmStats{Cities: len(cities)}
	for _, city := range cities {
		if _, err := s.Grid(ctx, city.Name, cellMeters); err != nil {
			return stats, err
		}
		if _, err := s.Recommendations(ctx, city.Name, recLimit); err != nil {
			return stats, err
		}
	}
	return stats, nil
}
