package coverage

import (
	"context"
	"testing"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
)

func TestIsExcludedCellRespectsHoles(t *testing.T) {
	areas := []domain.ExclusionArea{
		{
			Bounds: domain.BoundingBox{MinLat: 0, MinLng: 0, MaxLat: 10, MaxLng: 10},
			OuterRings: [][]domain.LatLng{{
				{Lat: 0, Lng: 0},
				{Lat: 0, Lng: 10},
				{Lat: 10, Lng: 10},
				{Lat: 10, Lng: 0},
				{Lat: 0, Lng: 0},
			}},
			InnerRings: [][]domain.LatLng{{
				{Lat: 4, Lng: 4},
				{Lat: 4, Lng: 6},
				{Lat: 6, Lng: 6},
				{Lat: 6, Lng: 4},
				{Lat: 4, Lng: 4},
			}},
		},
	}

	if !isExcludedCell(1, 1, areas) {
		t.Fatal("expected point inside outer ring to be excluded")
	}
	if isExcludedCell(5, 5, areas) {
		t.Fatal("expected point inside inner hole to stay included")
	}
	if isExcludedCell(11, 11, areas) {
		t.Fatal("expected point outside bounds to stay included")
	}
}

func TestBuildGridDropsExcludedTerrain(t *testing.T) {
	svc := &Service{
		Points: stubPointRepo{
			points: []domain.CoveragePoint{{ID: 1, Latitude: 50, Longitude: 20}},
			bbox:   domain.BoundingBox{MinLat: 50, MinLng: 20, MaxLat: 50, MaxLng: 20},
		},
		Competitors: stubCompetitorRepo{},
		Anchors:     stubAnchorRepo{},
		Terrain: stubTerrainClient{
			areas: []domain.ExclusionArea{{
				Bounds: domain.BoundingBox{MinLat: 49.98, MinLng: 19.98, MaxLat: 50.02, MaxLng: 20.02},
				OuterRings: [][]domain.LatLng{{
					{Lat: 49.98, Lng: 19.98},
					{Lat: 49.98, Lng: 20.02},
					{Lat: 50.02, Lng: 20.02},
					{Lat: 50.02, Lng: 19.98},
					{Lat: 49.98, Lng: 19.98},
				}},
			}},
		},
	}

	cells, summary, err := svc.BuildGrid(context.Background(), "Test", 300)
	if err != nil {
		t.Fatalf("BuildGrid error = %v", err)
	}
	if len(cells) != 0 {
		t.Fatalf("cells = %d, want 0", len(cells))
	}
	if summary.TotalCells == 0 {
		t.Fatal("expected grid to evaluate at least one cell")
	}
	if summary.GreenfieldCells != 0 || summary.CompetitiveCells != 0 || summary.InpostOnlyCells != 0 || summary.SaturatedCells != 0 {
		t.Fatalf("expected all scored cells to be excluded, got summary %+v", summary)
	}
}

type stubPointRepo struct {
	points []domain.CoveragePoint
	bbox   domain.BoundingBox
}

func (s stubPointRepo) Upsert(context.Context, *domain.Point) error { return nil }
func (s stubPointRepo) AllForCoverage(context.Context, domain.BoundingBox) ([]domain.CoveragePoint, error) {
	return s.points, nil
}
func (s stubPointRepo) BoundingBox(context.Context, string) (domain.BoundingBox, error) {
	return s.bbox, nil
}
func (s stubPointRepo) ListProvinces(context.Context, int) ([]domain.ProvinceInfo, error) {
	return nil, nil
}

type stubCompetitorRepo struct{}

func (stubCompetitorRepo) UpsertBatch(context.Context, []domain.CompetitorPoint) error { return nil }
func (stubCompetitorRepo) AllForCoverage(context.Context, domain.BoundingBox) ([]domain.CompetitorPoint, error) {
	return nil, nil
}
func (stubCompetitorRepo) Count(context.Context) (int, error) { return 0, nil }

type stubAnchorRepo struct{}

func (stubAnchorRepo) UpsertBatch(context.Context, []domain.AnchorPOI) error { return nil }
func (stubAnchorRepo) All(context.Context, domain.BoundingBox) ([]domain.AnchorPOI, error) {
	return nil, nil
}
func (stubAnchorRepo) Count(context.Context) (int, error) { return 0, nil }

type stubTerrainClient struct {
	areas []domain.ExclusionArea
}

func (s stubTerrainClient) FetchExclusionAreas(context.Context, domain.BoundingBox) ([]domain.ExclusionArea, error) {
	return s.areas, nil
}
