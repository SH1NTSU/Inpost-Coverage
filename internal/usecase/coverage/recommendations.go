package coverage

import (
	"context"
	"fmt"
	"sort"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
	"github.com/marcelbudziszewski/paczkomat-predictor/pkg/geo"
	"github.com/marcelbudziszewski/paczkomat-predictor/pkg/spatial"
)

// noNearbySentinel marks "no point of this kind found in the index".
// Anything at or above noNearbyThreshold should be treated as missing by callers.
const (
	noNearbySentinel  = 1e12
	noNearbyThreshold = 1e11
)

// ProvinceCompetitors returns competitor lockers that belong to the selected
// province by Voronoi-on-lockers — a competitor whose nearest InPost locker
// is tagged in another province is "their" competitor, not ours. The previous
// no-filter version bled neighbouring cities (Łódź into mazowieckie's view)
// because province bboxes overlap. The binary Tag check keeps almost all
// rural points (they're still nearest to a same-province locker) and only
// drops the cross-border cases.
func (s *Service) ProvinceCompetitors(ctx context.Context, province string) ([]domain.CompetitorPoint, error) {
	bb, err := s.resolveBBox(ctx, province)
	if err != nil {
		return nil, err
	}
	comps, err := s.Competitors.AllForCoverage(ctx, bb)
	if err != nil {
		return nil, err
	}
	if province == "" || bb.IsZero() {
		return comps, nil
	}
	inp, err := s.Points.AllForCoverage(ctx, bb)
	if err != nil {
		return nil, err
	}
	inpIdx := indexFromPoints(inp)
	filtered := comps[:0]
	for _, c := range comps {
		inP, _, ok := inpIdx.Nearest(c.Latitude, c.Longitude)
		if ok && inP.Tag != "" && inP.Tag != province {
			continue
		}
		filtered = append(filtered, c)
	}
	return filtered, nil
}

func (s *Service) Recommendations(ctx context.Context, province string, limit int) (domain.CoverageRecommendations, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	cacheKey := fmt.Sprintf("coverage:recs:%s:%s:%d", recommendationsCacheVersion, province, limit)
	var cached domain.CoverageRecommendations
	if s.Cache != nil {
		if hit, _ := s.Cache.Get(ctx, cacheKey, &cached); hit {
			return cached, nil
		}
	}
	if s.Store != nil {
		if snap, err := s.Store.LoadRecommendations(ctx, province, limit, recommendationsCacheVersion); err == nil && snap != nil {
			if s.Cache != nil {
				_ = s.Cache.Set(ctx, cacheKey, *snap)
			}
			return *snap, nil
		}
	}

	bb, err := s.resolveBBox(ctx, province)
	if err != nil {
		return domain.CoverageRecommendations{}, err
	}
	inp, err := s.Points.AllForCoverage(ctx, bb)
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
	// Province-only InPost index used as a foreign-territory gate so we don't
	// recommend villages on the wrong side of the country border.
	provLockers := make([]domain.CoveragePoint, 0, len(inp))
	for _, p := range inp {
		if province == "" || p.Province == province {
			provLockers = append(provLockers, p)
		}
	}
	inpProvinceIdx := indexFromPoints(provLockers)
	cpIdx := indexFromCompetitors(comps)

	// Three anchor indexes. Commercial anchors (shops, fuel, transit, schools,
	// post offices, etc.) feed snap targets, nearby-context, and a 1.5 km
	// habitability halo. Settlement markers (place=village/hamlet/town) feed
	// habitability only, with a much tighter halo because the node is a label
	// at the village centroid, not a 1.5 km service area. Without this split
	// the entire province lights up as inhabited because villages are
	// everywhere in OSM.
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

	newPoints := underservedSuggestions(province, bb, commercial, inpIdx, inpProvinceIdx, cpIdx, commercialIdx, settlementIdx, limit)
	upgrades := s.upgradeCandidatesWith(ctx, province, cpIdx, inp, limit)

	out := domain.CoverageRecommendations{
		NewPoints: newPoints,
		Upgrades:  upgrades,
	}
	if s.Store != nil {
		_ = s.Store.SaveRecommendations(ctx, province, limit, recommendationsCacheVersion, out)
	}
	if s.Cache != nil {
		_ = s.Cache.Set(ctx, cacheKey, out)
	}
	return out, nil
}

// underservedSuggestions walks the bbox on a coarse grid and emits one
// candidate per habitable underserved spot. If an OSM anchor sits within
// anchorSnapM of the scan point the candidate pins to that anchor (and is
// scored by anchor type); otherwise the candidate is an "open area" the
// operator could place a locker on a private/host site. Candidates are
// deduplicated by minSpacingM so we never return two pins on top of each
// other.
func underservedSuggestions(
	province string,
	bb domain.BoundingBox,
	commercialAnchors []domain.AnchorPOI,
	inpIdx, inpProvinceIdx, cpIdx, commercialIdx, settlementIdx *spatial.Index,
	limit int,
) []domain.GapSuggestion {
	if bb.IsZero() || limit <= 0 {
		return nil
	}

	const (
		scanStepM      = 1000.0
		anchorSnapM    = 250.0
		habitableM     = 1500.0 // commercial / competitor halo (settlements use settlementHabitableRadiusM per type)
		minSpacingM    = 1500.0
		openAreaWeight = 0.85
		nearbyContextM = 250.0
		nearbyContextN = 4
	)

	stepLat := scanStepM / metersPerDegLat
	stepLng := scanStepM / metersPerDegLng(50.06)

	// Index uses the DB primary key (a.ID) as Point.ID — keep this map
	// aligned with that, or the snap lookup silently misses every time and
	// every candidate falls through to "open area".
	byID := make(map[int64]domain.AnchorPOI, len(commercialAnchors))
	for _, a := range commercialAnchors {
		byID[a.ID] = a
	}

	type cand struct {
		lat, lng float64
		inDist   float64
		cpDist   float64
		cpNet    string
		tier     domain.CoverageTier
		anchor   *domain.AnchorPOI
		score    float64
	}

	var cands []cand
	for lat := bb.MinLat; lat <= bb.MaxLat; lat += stepLat {
		for lng := bb.MinLng; lng <= bb.MaxLng; lng += stepLng {
			inP, inDist, inOk := inpIdx.Nearest(lat, lng)
			// Voronoi-on-lockers province containment.
			if province != "" && inOk && inP.Tag != "" && inP.Tag != province {
				continue
			}
			// Foreign-territory gate — see grid.go for the rationale.
			if province != "" {
				_, provDist, provOk := inpProvinceIdx.Nearest(lat, lng)
				if !provOk || provDist > 6000 {
					continue
				}
			}
			if !inOk {
				inDist = noNearbySentinel
			}
			if inDist <= thresholdPoor {
				continue
			}
			// Habitability: commercial / competitor within 1.5 km, OR a
			// real settlement (town 1.5 km halo, village 0.5 km halo,
			// hamlets ignored — they're tagging noise).
			habitable := commercialIdx.HasWithin(lat, lng, habitableM) ||
				cpIdx.HasWithin(lat, lng, habitableM) ||
				settlementHabitable(lat, lng, settlementIdx)
			if !habitable {
				continue
			}

			cpP, cpDist, cpOk := cpIdx.Nearest(lat, lng)
			if !cpOk {
				cpDist = noNearbySentinel
			}
			cpNet := ""
			if cpOk {
				cpNet = cpP.Tag
			}
			tier := classify(inDist, cpDist)

			spotLat, spotLng := lat, lng
			weight := openAreaWeight
			var ap *domain.AnchorPOI
			// Snap to a commercial anchor only — never to a village name
			// marker. If only a settlement marker is nearby, the candidate
			// stays as an "open area" so the operator picks a host site.
			if anP, anDist, ok := commercialIdx.Nearest(lat, lng); ok && anDist <= anchorSnapM {
				if a, found := byID[anP.ID]; found {
					ap = &a
					spotLat, spotLng = a.Latitude, a.Longitude
					weight = anchorTypeWeight(a.Type)
				}
			}
			cands = append(cands, cand{
				lat: spotLat, lng: spotLng,
				inDist: inDist, cpDist: cpDist, cpNet: cpNet,
				tier: tier, anchor: ap, score: inDist * weight,
			})
		}
	}

	// Split candidates by kind so each group gets a fair share of slots.
	// A flat anchor-first sort would let mazowieckie's thousands of shops
	// fill every slot before a single open-area candidate is considered.
	var anchored, open []cand
	for _, c := range cands {
		if c.anchor != nil {
			anchored = append(anchored, c)
		} else {
			open = append(open, c)
		}
	}
	sort.Slice(anchored, func(i, j int) bool { return anchored[i].score > anchored[j].score })
	sort.Slice(open, func(i, j int) bool { return open[i].score > open[j].score })

	// Quota: 70 % anchored / 30 % open. Anchored show up first in the
	// returned list (filter chips on the frontend make this irrelevant for
	// users who only want one kind). If a category runs short, the other
	// fills the leftover slots so we never silently truncate below `limit`.
	anchorQuota := (limit * 7) / 10
	if anchorQuota < 1 {
		anchorQuota = 1
	}

	tooCloseToAny := func(c cand, picked []cand) bool {
		for _, p := range picked {
			if geo.Haversine(c.lat, c.lng, p.lat, p.lng) < minSpacingM {
				return true
			}
		}
		return false
	}

	picked := make([]cand, 0, limit)
	for _, c := range anchored {
		if len(picked) >= anchorQuota {
			break
		}
		if tooCloseToAny(c, picked) {
			continue
		}
		picked = append(picked, c)
	}
	for _, c := range open {
		if len(picked) >= limit {
			break
		}
		if tooCloseToAny(c, picked) {
			continue
		}
		picked = append(picked, c)
	}
	// If anchored under-filled its quota, top up with more anchored picks
	// (skipping ones already dedup'd) before declaring the list complete.
	if len(picked) < limit {
		for _, c := range anchored {
			if len(picked) >= limit {
				break
			}
			if tooCloseToAny(c, picked) {
				continue
			}
			picked = append(picked, c)
		}
	}

	out := make([]domain.GapSuggestion, 0, len(picked))
	for _, c := range picked {
		var primary *domain.NearbyAnchor
		exclID := int64(0)
		if c.anchor != nil {
			primary = &domain.NearbyAnchor{
				Type:      c.anchor.Type,
				Brand:     c.anchor.Brand,
				Name:      c.anchor.Name,
				Latitude:  c.anchor.Latitude,
				Longitude: c.anchor.Longitude,
				DistanceM: 0,
			}
			// nearbyAnchorsFromIndex compares against spatial.Point.ID
			// which is the DB primary key, not the OSM id.
			exclID = c.anchor.ID
		}
		ctxAnchors := nearbyAnchorsFromIndex(c.lat, c.lng, commercialIdx, exclID, nearbyContextM, nearbyContextN)
		out = append(out, domain.GapSuggestion{
			Lat:                  c.lat,
			Lng:                  c.lng,
			Tier:                 c.tier,
			NearestInpostM:       c.inDist,
			NearestCompetitorM:   c.cpDist,
			NearestCompetitorNet: c.cpNet,
			Reason:               buildSuggestionReason(c.anchor, c.tier, c.inDist),
			Anchor:               primary,
			NearbyAnchors:        ctxAnchors,
		})
	}
	return out
}

func buildSuggestionReason(a *domain.AnchorPOI, tier domain.CoverageTier, inDist float64) string {
	if a != nil {
		return buildAnchorReason(*a, tier, inDist)
	}
	inPart := "no InPost in range"
	if inDist < noNearbyThreshold {
		inPart = fmt.Sprintf("%s to nearest InPost", formatMetersShort(inDist))
	}
	switch tier {
	case domain.TierGreenfield:
		return fmt.Sprintf("Open underserved area — %s, no competitor mapped nearby", inPart)
	case domain.TierCompetitive:
		return fmt.Sprintf("Open underserved area — %s, competitor already nearby", inPart)
	}
	return fmt.Sprintf("Open underserved area — %s", inPart)
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
	case "post_office":
		return 1.05
	case "university":
		return 1.0
	case "school":
		return 1.0
	case "pharmacy":
		return 1.0
	case "convenience":
		return 0.95
	case "marketplace":
		return 0.8
	}
	return 1.0
}

// isSettlementType reports whether an anchor is a place-marker node
// (place=village/hamlet/town etc.). Those count as evidence of human
// presence for habitability but are not real businesses, so we never
// pin a recommendation onto one — "put a paczkomat at the village
// centroid label" is not actionable.
func isSettlementType(t string) bool {
	switch t {
	case "town", "village", "hamlet":
		return true
	}
	return false
}

// settlementHabitableRadiusM returns the radius around a settlement
// node within which we trust there are actually people. Loosely
// modelled on OSM convention: town/city nodes typically mark 5k+
// inhabitants with several km of built area, villages are a few
// hundred meters across, hamlets are 5–20 houses — basically noise
// from a population standpoint, so we exclude them.
func settlementHabitableRadiusM(t string) float64 {
	switch t {
	case "town":
		return 1500
	case "village":
		return 500
	}
	return 0
}

// settlementHabitable returns true if at least one place-marker is
// within its type-specific halo. Replaces the previous flat 600 m
// halo, which over-counted hamlets (one OSM hamlet node = 5 houses
// but currently "covered" 1.13 km² of habitable area).
func settlementHabitable(lat, lng float64, idx *spatial.Index) bool {
	const maxRadius = 1500.0 // widest halo (town)
	hits := idx.Within(lat, lng, maxRadius)
	for _, p := range hits {
		r := settlementHabitableRadiusM(p.Tag)
		if r == 0 {
			continue
		}
		if geo.Haversine(lat, lng, p.Lat, p.Lng) <= r {
			return true
		}
	}
	return false
}

func buildAnchorReason(a domain.AnchorPOI, tier domain.CoverageTier, inDist float64) string {
	label := a.Brand
	if label == "" {
		label = a.Name
	}
	if label == "" {
		label = a.Type
	}
	inPart := "no InPost in range"
	if inDist < noNearbyThreshold {
		inPart = fmt.Sprintf("%s to nearest InPost", formatMetersShort(inDist))
	}
	switch tier {
	case domain.TierGreenfield:
		return fmt.Sprintf("%s with %s — no competitor mapped nearby", label, inPart)
	case domain.TierCompetitive:
		return fmt.Sprintf("%s with %s — competitor already nearby", label, inPart)
	}
	return fmt.Sprintf("%s — %s", label, inPart)
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

func (s *Service) upgradeCandidatesWith(ctx context.Context, province string, cpIdx *spatial.Index, pts []domain.CoveragePoint, limit int) []domain.UpgradeCandidate {
	full, err := s.Analytics.ListLockers(ctx, nil, province)
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
