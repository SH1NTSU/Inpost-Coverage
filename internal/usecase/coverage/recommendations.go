package coverage

import (
	"context"
	"fmt"
	"sort"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
	"github.com/marcelbudziszewski/paczkomat-predictor/pkg/geo"
	"github.com/marcelbudziszewski/paczkomat-predictor/pkg/spatial"
)

func (s *Service) Recommendations(ctx context.Context, city string, limit int) (domain.CoverageRecommendations, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	cacheKey := fmt.Sprintf("coverage:recs:%s:%s:%d", recommendationsCacheVersion, city, limit)
	var cached domain.CoverageRecommendations
	if s.Cache != nil {
		if hit, _ := s.Cache.Get(ctx, cacheKey, &cached); hit {
			return cached, nil
		}
	}
	if s.Store != nil {
		if snap, err := s.Store.LoadRecommendations(ctx, city, limit, recommendationsCacheVersion); err == nil && snap != nil {
			if s.Cache != nil {
				_ = s.Cache.Set(ctx, cacheKey, *snap)
			}
			return *snap, nil
		}
	}

	bb, err := s.resolveBBox(ctx, city)
	if err != nil {
		return domain.CoverageRecommendations{}, err
	}
	inp, err := s.Points.AllForCoverage(ctx, city)
	if err != nil {
		return domain.CoverageRecommendations{}, err
	}
	comps, err := s.Competitors.AllForCoverage(ctx, bb)
	if err != nil {
		return domain.CoverageRecommendations{}, err
	}
	anchors, err := s.Anchors.All(ctx, bb)
	if err != nil {
		return domain.CoverageRecommendations{}, err
	}

	inpIdx := indexFromPoints(inp)
	cpIdx := indexFromCompetitors(comps)
	anchorIdx := indexFromAnchors(anchors)

	newPoints := anchorDrivenSuggestions(anchors, inpIdx, cpIdx, anchorIdx, limit)
	upgrades := s.upgradeCandidatesWith(ctx, city, cpIdx, inp, limit)

	out := domain.CoverageRecommendations{
		NewPoints: newPoints,
		Upgrades:  upgrades,
	}
	if s.Store != nil {
		_ = s.Store.SaveRecommendations(ctx, city, limit, recommendationsCacheVersion, out)
	}
	if s.Cache != nil {
		_ = s.Cache.Set(ctx, cacheKey, out)
	}
	return out, nil
}

func anchorDrivenSuggestions(
	anchors []domain.AnchorPOI,
	inpIdx, cpIdx, anchorIdx *spatial.Index,
	limit int,
) []domain.GapSuggestion {
	type scored struct {
		a      domain.AnchorPOI
		inDist float64
		cpDist float64
		cpNet  string
		tier   domain.CoverageTier
		score  float64
	}

	cands := make([]scored, 0, len(anchors))
	for _, a := range anchors {
		_, inDist, inOk := inpIdx.Nearest(a.Latitude, a.Longitude)
		if !inOk {
			inDist = 1e12
		}
		if inDist <= thresholdPoor {
			continue
		}
		cpP, cpDist, cpOk := cpIdx.Nearest(a.Latitude, a.Longitude)
		if !cpOk {
			cpDist = 1e12
		}
		net := ""
		if cpOk {
			net = cpP.Tag
		}
		tier := classify(inDist, cpDist)
		score := inDist * anchorTypeWeight(a.Type)
		cands = append(cands, scored{a, inDist, cpDist, net, tier, score})
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].score > cands[j].score })
	if len(cands) > limit {
		cands = cands[:limit]
	}

	out := make([]domain.GapSuggestion, 0, len(cands))
	for _, c := range cands {
		primary := &domain.NearbyAnchor{
			Type:      c.a.Type,
			Brand:     c.a.Brand,
			Name:      c.a.Name,
			Latitude:  c.a.Latitude,
			Longitude: c.a.Longitude,
			DistanceM: 0,
		}
		context := nearbyAnchorsFromIndex(c.a.Latitude, c.a.Longitude, anchorIdx, c.a.OSMID, 250, 4)
		out = append(out, domain.GapSuggestion{
			Lat:                  c.a.Latitude,
			Lng:                  c.a.Longitude,
			Tier:                 c.tier,
			NearestInpostM:       c.inDist,
			NearestCompetitorM:   c.cpDist,
			NearestCompetitorNet: c.cpNet,
			Reason:               buildAnchorReason(c.a, c.tier, c.inDist),
			Anchor:               primary,
			NearbyAnchors:        context,
		})
	}
	return out
}

func anchorTypeWeight(t string) float64 {
	switch t {
	case "mall":
		return 1.4
	case "supermarket":
		return 1.2
	case "fuel":
		return 1.15
	case "transit":
		return 1.1
	case "university":
		return 1.0
	case "convenience":
		return 0.95
	case "marketplace":
		return 0.8
	}
	return 1.0
}

func buildAnchorReason(a domain.AnchorPOI, tier domain.CoverageTier, inDist float64) string {
	label := a.Brand
	if label == "" {
		label = a.Name
	}
	if label == "" {
		label = a.Type
	}
	switch tier {
	case domain.TierGreenfield:
		return fmt.Sprintf("%s with %s to nearest InPost — no competitor either",
			label, formatMetersShort(inDist))
	case domain.TierCompetitive:
		return fmt.Sprintf("%s with %s to nearest InPost — competitor already nearby",
			label, formatMetersShort(inDist))
	}
	return fmt.Sprintf("%s — %s to nearest InPost", label, formatMetersShort(inDist))
}

func formatMetersShort(m float64) string {
	if m < 1000 {
		return fmt.Sprintf("%.0f m", m)
	}
	return fmt.Sprintf("%.1f km", m/1000)
}

func nearbyAnchorsFromIndex(lat, lng float64, idx *spatial.Index, exclOSMID int64, radiusM float64, max int) []domain.NearbyAnchor {
	type scored struct {
		p spatial.Point
		d float64
	}
	hits := idx.Within(lat, lng, radiusM)
	cands := make([]scored, 0, len(hits))
	for _, p := range hits {
		if p.ID == exclOSMID {
			continue
		}
		cands = append(cands, scored{p, geo.Haversine(lat, lng, p.Lat, p.Lng)})
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].d < cands[j].d })
	if len(cands) > max {
		cands = cands[:max]
	}
	out := make([]domain.NearbyAnchor, len(cands))
	for i, c := range cands {
		out[i] = domain.NearbyAnchor{
			Type:      c.p.Tag,
			Latitude:  c.p.Lat,
			Longitude: c.p.Lng,
			DistanceM: c.d,
		}
	}
	return out
}

func (s *Service) upgradeCandidatesWith(ctx context.Context, city string, cpIdx *spatial.Index, pts []domain.CoveragePoint, limit int) []domain.UpgradeCandidate {
	full, err := s.Analytics.ListLockers(ctx, nil, city)
	if err != nil {
		return nil
	}
	bySummaryID := make(map[int64]domain.LockerSummary, len(full))
	for _, l := range full {
		bySummaryID[l.ID] = l
	}

	cands := make([]domain.UpgradeCandidate, 0, len(pts))
	for _, p := range pts {
		if p.IsNext {
			continue
		}
		pressure := len(cpIdx.Within(p.Latitude, p.Longitude, 400))
		score, reasons := scoreUpgrade(p.IsNext, pressure)
		if score <= 0 {
			continue
		}
		sum, ok := bySummaryID[p.ID]
		if !ok {
			continue
		}
		cands = append(cands, domain.UpgradeCandidate{
			Locker:             sum,
			IsNext:             p.IsNext,
			CompetitorPressure: pressure,
			Score:              score,
			Reasons:            reasons,
		})
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].Score > cands[j].Score })
	if len(cands) > limit {
		cands = cands[:limit]
	}
	return cands
}

func scoreUpgrade(isNext bool, pressure int) (float64, []string) {
	if isNext {
		return 0, nil
	}
	score := 0.5
	reasons := []string{"Old generation hardware"}
	if pressure > 0 {
		score += float64(pressure) * 0.05
		reasons = append(reasons,
			fmt.Sprintf("%d competitor locker(s) within 400 m", pressure))
	}
	return score, reasons
}
