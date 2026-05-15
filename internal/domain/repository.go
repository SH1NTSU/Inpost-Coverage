package domain

import "context"

type PointRepository interface {
	Upsert(ctx context.Context, p *Point) error

	AllForCoverage(ctx context.Context, bbox BoundingBox) ([]CoveragePoint, error)

	BoundingBox(ctx context.Context, province string) (BoundingBox, error)

	ListProvinces(ctx context.Context, minPoints int) ([]ProvinceInfo, error)
}

type ProvinceInfo struct {
	Name       string  `json:"name"`
	PointCount int     `json:"point_count"`
	MinLat     float64 `json:"min_lat"`
	MinLng     float64 `json:"min_lng"`
	MaxLat     float64 `json:"max_lat"`
	MaxLng     float64 `json:"max_lng"`
	CenterLat  float64 `json:"center_lat"`
	CenterLng  float64 `json:"center_lng"`
}

type SnapshotRepository interface {
	InsertBatch(ctx context.Context, items []AvailabilitySnapshot) error
}

type CompetitorRepository interface {
	UpsertBatch(ctx context.Context, items []CompetitorPoint) error

	AllForCoverage(ctx context.Context, bbox BoundingBox) ([]CompetitorPoint, error)
	Count(ctx context.Context) (int, error)
}

type AnchorRepository interface {
	UpsertBatch(ctx context.Context, items []AnchorPOI) error

	All(ctx context.Context, bbox BoundingBox) ([]AnchorPOI, error)
	Count(ctx context.Context) (int, error)
}

func (b BoundingBox) IsZero() bool {
	return b.MinLat == 0 && b.MinLng == 0 && b.MaxLat == 0 && b.MaxLng == 0
}

type CoveragePoint struct {
	ID        int64
	Latitude  float64
	Longitude float64
	IsNext    bool
	Province  string
}

type BoundingBox struct {
	MinLat, MinLng, MaxLat, MaxLng float64
}

type AnalyticsRepository interface {
	Stats(ctx context.Context) (Stats, error)
	ListLockers(ctx context.Context, statusFilter *PointStatus, province string) ([]LockerSummary, error)
	GetLocker(ctx context.Context, id int64) (*LockerDetail, error)
	History(ctx context.Context, id int64, limit int) ([]StateChange, error)
	CurrentOutages(ctx context.Context) ([]OngoingOutage, error)

	WorstOffenders(ctx context.Context, days, limit int) ([]WorstOffender, error)
	OutagesTimeline(ctx context.Context, days int) ([]DailyOutages, error)
	UptimeDistribution(ctx context.Context, days int) ([]UptimeBucket, error)
	NetworkStats(ctx context.Context, days int) (NetworkStats, error)

	LockerUptime7d(ctx context.Context) (map[int64]LockerUptime, error)
}

type CoverageCacheRepository interface {
	LoadGrid(ctx context.Context, province string, cellMeters int, version string) (*CoverageGridSnapshot, error)
	SaveGrid(ctx context.Context, province string, cellMeters int, version string, snap CoverageGridSnapshot) error
	LoadRecommendations(ctx context.Context, province string, limit int, version string) (*CoverageRecommendations, error)
	SaveRecommendations(ctx context.Context, province string, limit int, version string, recs CoverageRecommendations) error
}

type LockerUptime struct {
	UptimePct    float64
	OutageEvents int
}
