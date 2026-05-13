package coverage

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/cache"
	"github.com/marcelbudziszewski/paczkomat-predictor/pkg/spatial"
)

const (
	thresholdGood = 300
	thresholdOK   = 600
	thresholdPoor = 1000

	habitableRadiusM = 700

	cacheKeyVersion = "v3"

	recommendationsCacheVersion = "v1"
)

const (
	metersPerDegLat = 111_320.0
)

func metersPerDegLng(_ float64) float64 { return 71500.0 }

type Service struct {
	Points      domain.PointRepository
	Competitors domain.CompetitorRepository
	Anchors     domain.AnchorRepository
	Analytics   domain.AnalyticsRepository
	Terrain     domain.TerrainClient
	Store       domain.CoverageCacheRepository
	Cache       cache.Cache
}

func (s *Service) resolveBBox(ctx context.Context, city string) (domain.BoundingBox, error) {
	bb, err := s.Points.BoundingBox(ctx, city)
	if err != nil {
		return bb, err
	}
	if bb.MinLat == 0 && bb.MaxLat == 0 {
		return bb, nil
	}
	bb.MinLat -= 0.01
	bb.MinLng -= 0.015
	bb.MaxLat += 0.01
	bb.MaxLng += 0.015
	return bb, nil
}

func (s *Service) BuildGrid(ctx context.Context, city string, cellMeters int) ([]domain.GridCell, domain.CoverageSummary, error) {
	snap, err := s.Grid(ctx, city, cellMeters)
	if err != nil {
		return nil, domain.CoverageSummary{}, err
	}
	return snap.Cells, snap.Summary, nil
}

func (s *Service) Grid(ctx context.Context, city string, cellMeters int) (domain.CoverageGridSnapshot, error) {
	if cellMeters <= 0 {
		cellMeters = 300
	}

	cacheKey := fmt.Sprintf("coverage:grid:%s:%s:%d", cacheKeyVersion, city, cellMeters)
	var cached domain.CoverageGridSnapshot
	if s.Cache != nil {
		if hit, _ := s.Cache.Get(ctx, cacheKey, &cached); hit {
			return cached, nil
		}
	}
	if s.Store != nil {
		if snap, err := s.Store.LoadGrid(ctx, city, cellMeters, cacheKeyVersion); err == nil && snap != nil {
			if s.Cache != nil {
				_ = s.Cache.Set(ctx, cacheKey, *snap)
			}
			return *snap, nil
		}
	}

	bb, err := s.resolveBBox(ctx, city)
	if err != nil {
		return domain.CoverageGridSnapshot{}, err
	}
	inp, err := s.Points.AllForCoverage(ctx, city)
	if err != nil {
		return domain.CoverageGridSnapshot{}, err
	}
	comps, err := s.Competitors.AllForCoverage(ctx, bb)
	if err != nil {
		return domain.CoverageGridSnapshot{}, err
	}

	anchors, err := s.Anchors.All(ctx, bb)
	if err != nil {
		return domain.CoverageGridSnapshot{}, err
	}
	exclusions := s.exclusionAreas(ctx, city, bb)

	inpIdx := indexFromPoints(inp)
	cpIdx := indexFromCompetitors(comps)
	anchorIdx := indexFromAnchors(anchors)

	stepLat := float64(cellMeters) / metersPerDegLat
	stepLng := float64(cellMeters) / metersPerDegLng(50.06)

	var rows []float64
	for lat := bb.MinLat; lat <= bb.MaxLat; lat += stepLat {
		rows = append(rows, lat)
	}

	type rowAcc struct {
		cells       []domain.GridCell
		green       int
		comp        int
		inpOnly     int
		sat         int
		uninhabited int
		total       int
	}

	workers := runtime.NumCPU()
	if workers > len(rows) {
		workers = len(rows)
	}
	if workers < 1 {
		workers = 1
	}
	rowResults := make([]rowAcc, len(rows))

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(worker int) {
			defer wg.Done()
			for i := worker; i < len(rows); i += workers {
				lat := rows[i]
				var acc rowAcc
				for lng := bb.MinLng; lng <= bb.MaxLng; lng += stepLng {
					centerLat := lat + stepLat/2
					centerLng := lng + stepLng/2

					if isExcludedCell(centerLat, centerLng, exclusions) {
						acc.total++
						acc.uninhabited++
						continue
					}

					inP, inDist, inOk := inpIdx.Nearest(centerLat, centerLng)
					cpP, cpDist, cpOk := cpIdx.Nearest(centerLat, centerLng)
					if !inOk {
						inDist = 1e12
					}
					if !cpOk {
						cpDist = 1e12
					}

					hasNetwork := inDist <= habitableRadiusM || cpDist <= habitableRadiusM
					if !hasNetwork && !anchorIdx.HasWithin(centerLat, centerLng, habitableRadiusM) {
						acc.total++
						acc.uninhabited++
						continue
					}

					tier := classify(inDist, cpDist)
					acc.total++
					switch tier {
					case domain.TierGreenfield:
						acc.green++
					case domain.TierCompetitive:
						acc.comp++
					case domain.TierInpostOnly:
						acc.inpOnly++
					case domain.TierSaturated:
						acc.sat++
					}

					if inDist <= thresholdGood {
						continue
					}

					cell := domain.GridCell{
						Lat:                centerLat,
						Lng:                centerLng,
						NearestInpostM:     inDist,
						NearestCompetitorM: cpDist,
						Tier:               tier,
					}
					if inOk {
						id := inP.ID
						cell.NearestInpostID = &id
					}
					if cpOk {
						cell.NearestCompetitorNet = cpP.Tag
					}
					acc.cells = append(acc.cells, cell)
				}
				rowResults[i] = acc
			}
		}(w)
	}
	wg.Wait()

	summary := domain.CoverageSummary{
		CellMeters:        cellMeters,
		InpostLockers:     len(inp),
		CompetitorLockers: len(comps),
	}
	var cells []domain.GridCell
	for _, acc := range rowResults {
		summary.TotalCells += acc.total
		summary.GreenfieldCells += acc.green
		summary.CompetitiveCells += acc.comp
		summary.InpostOnlyCells += acc.inpOnly
		summary.SaturatedCells += acc.sat
		cells = append(cells, acc.cells...)
	}
	cellAreaKm2 := float64(cellMeters*cellMeters) / 1_000_000.0
	underservedCells := summary.GreenfieldCells + summary.CompetitiveCells
	summary.UnderservedKm2 = float64(underservedCells) * cellAreaKm2

	snap := domain.CoverageGridSnapshot{
		Summary: summary,
		Cells:   cells,
	}
	if s.Store != nil {
		_ = s.Store.SaveGrid(ctx, city, cellMeters, cacheKeyVersion, snap)
	}
	if s.Cache != nil {
		_ = s.Cache.Set(ctx, cacheKey, snap)
	}
	return snap, nil
}

func classify(inDist, cpDist float64) domain.CoverageTier {
	inNear := inDist <= thresholdPoor
	cpNear := cpDist <= thresholdPoor
	switch {
	case inNear && cpNear:
		return domain.TierSaturated
	case inNear:
		return domain.TierInpostOnly
	case cpNear:
		return domain.TierCompetitive
	default:
		return domain.TierGreenfield
	}
}

func indexFromPoints(pts []domain.CoveragePoint) *spatial.Index {
	out := make([]spatial.Point, len(pts))
	for i, p := range pts {
		out[i] = spatial.Point{Lat: p.Latitude, Lng: p.Longitude, ID: p.ID}
	}
	return spatial.New(out, 500)
}

func indexFromCompetitors(pts []domain.CompetitorPoint) *spatial.Index {
	out := make([]spatial.Point, len(pts))
	for i, p := range pts {
		out[i] = spatial.Point{Lat: p.Latitude, Lng: p.Longitude, ID: p.ID, Tag: p.Network}
	}
	return spatial.New(out, 500)
}

func indexFromAnchors(pts []domain.AnchorPOI) *spatial.Index {
	out := make([]spatial.Point, len(pts))
	for i, p := range pts {
		out[i] = spatial.Point{Lat: p.Latitude, Lng: p.Longitude, ID: p.ID, Tag: p.Type}
	}
	return spatial.New(out, 500)
}
