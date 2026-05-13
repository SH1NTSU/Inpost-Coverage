package domain

import "context"

type InpostClient interface {
	ListByCity(ctx context.Context, country, city string, pageSize int,
		yield func(page []Point, snapshots []AvailabilitySnapshot) error) error
}

type TerrainClient interface {
	FetchExclusionAreas(ctx context.Context, bbox BoundingBox) ([]ExclusionArea, error)
}
