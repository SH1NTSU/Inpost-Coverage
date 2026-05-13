package overpass

import "testing"

func TestExclusionAreaFromRelationAssemblesRings(t *testing.T) {
	area, ok := exclusionAreaFromElement(overpassElement{
		Type: "relation",
		Tags: map[string]string{"natural": "water"},
		Members: []overpassMember{
			{
				Role: "outer",
				Geometry: []overpassCoord{
					{Lat: 0, Lon: 0},
					{Lat: 0, Lon: 4},
					{Lat: 4, Lon: 4},
				},
			},
			{
				Role: "outer",
				Geometry: []overpassCoord{
					{Lat: 4, Lon: 4},
					{Lat: 4, Lon: 0},
					{Lat: 0, Lon: 0},
				},
			},
			{
				Role: "inner",
				Geometry: []overpassCoord{
					{Lat: 1, Lon: 1},
					{Lat: 1, Lon: 2},
					{Lat: 2, Lon: 2},
					{Lat: 2, Lon: 1},
					{Lat: 1, Lon: 1},
				},
			},
		},
	})
	if !ok {
		t.Fatal("expected relation area to be built")
	}
	if area.Category != "water" {
		t.Fatalf("category = %q, want water", area.Category)
	}
	if got := len(area.OuterRings); got != 1 {
		t.Fatalf("outer ring count = %d, want 1", got)
	}
	if got := len(area.InnerRings); got != 1 {
		t.Fatalf("inner ring count = %d, want 1", got)
	}
	if area.Bounds.MinLat != 0 || area.Bounds.MinLng != 0 || area.Bounds.MaxLat != 4 || area.Bounds.MaxLng != 4 {
		t.Fatalf("unexpected bounds: %+v", area.Bounds)
	}
}
