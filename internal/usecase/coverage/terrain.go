package coverage

import (
	"context"
	"fmt"
	"math"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
)

const terrainCacheVersion = "v1"

func (s *Service) exclusionAreas(ctx context.Context, province string, bb domain.BoundingBox) []domain.ExclusionArea {
	if s.Terrain == nil || bb.IsZero() {
		return nil
	}

	cacheKey := fmt.Sprintf("coverage:terrain:%s:%s:%.5f:%.5f:%.5f:%.5f",
		terrainCacheVersion, province, bb.MinLat, bb.MinLng, bb.MaxLat, bb.MaxLng)
	var cached []domain.ExclusionArea
	if s.Cache != nil {
		if hit, _ := s.Cache.Get(ctx, cacheKey, &cached); hit {
			return cached
		}
	}

	areas, err := s.Terrain.FetchExclusionAreas(ctx, bb)
	if err != nil {
		return nil
	}
	if s.Cache != nil {
		_ = s.Cache.Set(ctx, cacheKey, areas)
	}
	return areas
}

func isExcludedCell(lat, lng float64, areas []domain.ExclusionArea) bool {
	for _, area := range areas {
		if lat < area.Bounds.MinLat || lat > area.Bounds.MaxLat ||
			lng < area.Bounds.MinLng || lng > area.Bounds.MaxLng {
			continue
		}
		if !pointInAnyRing(lat, lng, area.OuterRings) {
			continue
		}
		if pointInAnyRing(lat, lng, area.InnerRings) {
			continue
		}
		return true
	}
	return false
}

func pointInAnyRing(lat, lng float64, rings [][]domain.LatLng) bool {
	for _, ring := range rings {
		if pointInRing(lat, lng, ring) {
			return true
		}
	}
	return false
}

func pointInRing(lat, lng float64, ring []domain.LatLng) bool {
	if len(ring) < 4 {
		return false
	}

	inside := false
	for i, j := 0, len(ring)-1; i < len(ring); j, i = i, i+1 {
		a := ring[j]
		b := ring[i]
		if pointOnSegment(lat, lng, a, b) {
			return true
		}
		intersects := ((a.Lat > lat) != (b.Lat > lat)) &&
			(lng <= (b.Lng-a.Lng)*(lat-a.Lat)/(b.Lat-a.Lat)+a.Lng)
		if intersects {
			inside = !inside
		}
	}
	return inside
}

func pointOnSegment(lat, lng float64, a, b domain.LatLng) bool {
	const eps = 1e-9

	cross := (lng-a.Lng)*(b.Lat-a.Lat) - (lat-a.Lat)*(b.Lng-a.Lng)
	if math.Abs(cross) > eps {
		return false
	}

	minLat := math.Min(a.Lat, b.Lat) - eps
	maxLat := math.Max(a.Lat, b.Lat) + eps
	minLng := math.Min(a.Lng, b.Lng) - eps
	maxLng := math.Max(a.Lng, b.Lng) + eps
	return lat >= minLat && lat <= maxLat && lng >= minLng && lng <= maxLng
}
