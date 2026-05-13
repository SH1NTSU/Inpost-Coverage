package domain

import "time"

type CompetitorPoint struct {
	ID        int64     `json:"id"`
	Network   string    `json:"network"`
	Name      string    `json:"name"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Address   string    `json:"address"`
	OSMID     int64     `json:"osm_id"`
	FetchedAt time.Time `json:"fetched_at"`
}

type CoverageTier string

const (
	TierGreenfield  CoverageTier = "greenfield"
	TierCompetitive CoverageTier = "competitive"
	TierInpostOnly  CoverageTier = "inpost_only"
	TierSaturated   CoverageTier = "saturated"
)

type GridCell struct {
	Lat                  float64      `json:"lat"`
	Lng                  float64      `json:"lng"`
	NearestInpostM       float64      `json:"nearest_inpost_m"`
	NearestInpostID      *int64       `json:"nearest_inpost_id,omitempty"`
	NearestCompetitorM   float64      `json:"nearest_competitor_m"`
	NearestCompetitorNet string       `json:"nearest_competitor_network,omitempty"`
	Tier                 CoverageTier `json:"tier"`
}

type CoverageSummary struct {
	CellMeters        int     `json:"cell_meters"`
	TotalCells        int     `json:"total_cells"`
	GreenfieldCells   int     `json:"greenfield_cells"`
	CompetitiveCells  int     `json:"competitive_cells"`
	InpostOnlyCells   int     `json:"inpost_only_cells"`
	SaturatedCells    int     `json:"saturated_cells"`
	UnderservedKm2    float64 `json:"underserved_km2"`
	InpostLockers     int     `json:"inpost_lockers"`
	CompetitorLockers int     `json:"competitor_lockers"`
}

type CoverageGridSnapshot struct {
	Summary CoverageSummary `json:"summary"`
	Cells   []GridCell      `json:"cells"`
}

type AnchorPOI struct {
	ID        int64     `json:"id"`
	Type      string    `json:"poi_type"`
	Brand     string    `json:"brand"`
	Name      string    `json:"name"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Address   string    `json:"address"`
	OSMID     int64     `json:"osm_id"`
	FetchedAt time.Time `json:"fetched_at"`
}

type NearbyAnchor struct {
	Type      string  `json:"poi_type"`
	Brand     string  `json:"brand"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	DistanceM float64 `json:"distance_m"`
}

type GapSuggestion struct {
	Lat                  float64      `json:"lat"`
	Lng                  float64      `json:"lng"`
	Tier                 CoverageTier `json:"tier"`
	NearestInpostM       float64      `json:"nearest_inpost_m"`
	NearestCompetitorM   float64      `json:"nearest_competitor_m"`
	NearestCompetitorNet string       `json:"nearest_competitor_network,omitempty"`
	Reason               string       `json:"reason"`

	Anchor *NearbyAnchor `json:"anchor"`

	NearbyAnchors []NearbyAnchor `json:"nearby_anchors"`
}

type UpgradeCandidate struct {
	Locker             LockerSummary `json:"locker"`
	IsNext             bool          `json:"is_next"`
	CompetitorPressure int           `json:"competitor_pressure"`
	Score              float64       `json:"score"`
	Reasons            []string      `json:"reasons"`
}

type CoverageRecommendations struct {
	NewPoints []GapSuggestion    `json:"new_points"`
	Upgrades  []UpgradeCandidate `json:"upgrades"`
}
