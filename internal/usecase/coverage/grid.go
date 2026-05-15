package coverage

import (
	"context"
	"fmt"
	"runtime"
	"sort"
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

	cacheKeyVersion = "v16-competitor-rescrape"

	recommendationsCacheVersion = "v18-competitor-rescrape"
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

// Approximate Polish national bounds, used as a hard cap so a single bad
// row in the points table (a locker with mangled lat/lng) can't drag a
// province's bbox into Belarus or the Czech Republic.
const (
	polandMinLat = 49.00
	polandMaxLat = 54.85
	polandMinLng = 14.10
	polandMaxLng = 24.15
)

func (s *Service) resolveBBox(ctx context.Context, province string) (domain.BoundingBox, error) {
	bb, err := s.Points.BoundingBox(ctx, province)
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
	if bb.MinLat < polandMinLat {
		bb.MinLat = polandMinLat
	}
	if bb.MaxLat > polandMaxLat {
		bb.MaxLat = polandMaxLat
	}
	if bb.MinLng < polandMinLng {
		bb.MinLng = polandMinLng
	}
	if bb.MaxLng > polandMaxLng {
		bb.MaxLng = polandMaxLng
	}
	return bb, nil
}

func (s *Service) BuildGrid(ctx context.Context, province string, cellMeters int) ([]domain.GridCell, domain.CoverageSummary, error) {
	snap, err := s.Grid(ctx, province, cellMeters)
	if err != nil {
		return nil, domain.CoverageSummary{}, err
	}
	return snap.Cells, snap.Summary, nil
}

func (s *Service) Grid(ctx context.Context, province string, cellMeters int) (domain.CoverageGridSnapshot, error) {
	if cellMeters <= 0 {
		cellMeters = 300
	}

	cacheKey := fmt.Sprintf("coverage:grid:%s:%s:%d", cacheKeyVersion, province, cellMeters)
	var cached domain.CoverageGridSnapshot
	if s.Cache != nil {
		if hit, _ := s.Cache.Get(ctx, cacheKey, &cached); hit {
			return cached, nil
		}
	}
	if s.Store != nil {
		if snap, err := s.Store.LoadGrid(ctx, province, cellMeters, cacheKeyVersion); err == nil && snap != nil {
			if s.Cache != nil {
				_ = s.Cache.Set(ctx, cacheKey, *snap)
			}
			return *snap, nil
		}
	}

	bb, err := s.resolveBBox(ctx, province)
	if err != nil {
		return domain.CoverageGridSnapshot{}, err
	}
	inp, err := s.Points.AllForCoverage(ctx, bb)
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
	exclusions := s.exclusionAreas(ctx, province, bb)

	inpIdx := indexFromPoints(inp)
	// Second InPost index containing only same-province lockers. Used as a
	// "must be within reach of the province" gate so cells over the Czech /
	// Slovak / Belarusian / German side of the bbox get dropped even when
	// Voronoi-by-locker can't distinguish them (a Polish locker 4 km from the
	// border is still the closest thing on both sides of that border).
	provLockers := make([]domain.CoveragePoint, 0, len(inp))
	for _, p := range inp {
		if province == "" || p.Province == province {
			provLockers = append(provLockers, p)
		}
	}
	inpProvinceIdx := indexFromPoints(provLockers)
	cpIdx := indexFromCompetitors(comps)
	// Split anchors into commercial vs settlement so the two get different
	// habitability radii — see recommendations.go for the rationale.
	commercial := make([]domain.AnchorPOI, 0, len(anchors))
	settlement := make([]domain.AnchorPOI, 0, len(anchors))
	for _, a := range anchors {
		if isSettlementType(a.Type) {
			settlement = append(settlement, a)
			continue
		}
		commercial = append(commercial, a)
	}
	commercialIdx := indexFromAnchors(commercial)
	settlementIdx := indexFromAnchors(settlement)

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
					// Voronoi-on-lockers province containment: a cell whose
					// nearest InPost is in a different province belongs to
					// that province's catchment — skip it so the choropleth
					// doesn't bleed across province borders.
					if province != "" && inOk && inP.Tag != "" && inP.Tag != province {
						continue
					}
					// Foreign-territory gate: real Polish cells almost always
					// have a same-province locker within ~6 km. Cells beyond
					// that are either in another country (Czech / Belarus) or
					// in totally unsettled wilderness — neither is meaningful.
					if province != "" {
						_, provDist, provOk := inpProvinceIdx.Nearest(centerLat, centerLng)
						if !provOk || provDist > 6000 {
							continue
						}
					}
					if !inOk {
						inDist = noNearbySentinel
					}
					if !cpOk {
						cpDist = noNearbySentinel
					}

					// Habitability mirror of recommendations.go: commercial /
					// competitor within 1.5 km, OR a real settlement node
					// (town 1.5 km, village 0.5 km, hamlets ignored — they're
					// 5-house clusters and would paint farmland otherwise).
					habitable := inDist <= habitableRadiusM ||
						cpDist <= habitableRadiusM ||
						commercialIdx.HasWithin(centerLat, centerLng, 1500) ||
						cpIdx.HasWithin(centerLat, centerLng, 1500) ||
						settlementHabitable(centerLat, centerLng, settlementIdx)
					if !habitable {
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
	// Collect emitted cells split by tier so we can cap how many of each go
	// over the wire. Stats above (acc.* counters) are unaffected, so the
	// "Underserved area" stat still reflects the *true* count.
	var greenCells, compCells, otherCells []domain.GridCell
	for _, acc := range rowResults {
		summary.TotalCells += acc.total
		summary.GreenfieldCells += acc.green
		summary.CompetitiveCells += acc.comp
		summary.InpostOnlyCells += acc.inpOnly
		summary.SaturatedCells += acc.sat
		for _, c := range acc.cells {
			switch c.Tier {
			case domain.TierGreenfield:
				greenCells = append(greenCells, c)
			case domain.TierCompetitive:
				compCells = append(compCells, c)
			default:
				otherCells = append(otherCells, c)
			}
		}
	}
	// Sort each tier by inDist descending so the worst gaps are kept when we
	// trim. The choropleth still looks right (densest red where the gaps
	// really are) and the browser doesn't choke on 20k+ rectangles.
	const maxGreenfieldEmit = 1500
	const maxCompetitiveEmit = 500
	sort.Slice(greenCells, func(i, j int) bool { return greenCells[i].NearestInpostM > greenCells[j].NearestInpostM })
	sort.Slice(compCells, func(i, j int) bool { return compCells[i].NearestInpostM > compCells[j].NearestInpostM })
	if len(greenCells) > maxGreenfieldEmit {
		greenCells = greenCells[:maxGreenfieldEmit]
	}
	if len(compCells) > maxCompetitiveEmit {
		compCells = compCells[:maxCompetitiveEmit]
	}
	cells := make([]domain.GridCell, 0, len(greenCells)+len(compCells)+len(otherCells))
	cells = append(cells, greenCells...)
	cells = append(cells, compCells...)
	cells = append(cells, otherCells...)

	cellAreaKm2 := float64(cellMeters*cellMeters) / 1_000_000.0
	underservedCells := summary.GreenfieldCells + summary.CompetitiveCells
	summary.UnderservedKm2 = float64(underservedCells) * cellAreaKm2

	snap := domain.CoverageGridSnapshot{
		Summary: summary,
		Cells:   cells,
	}
	if s.Store != nil {
		_ = s.Store.SaveGrid(ctx, province, cellMeters, cacheKeyVersion, snap)
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
		out[i] = spatial.Point{Lat: p.Latitude, Lng: p.Longitude, ID: p.ID, Tag: p.Province}
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
