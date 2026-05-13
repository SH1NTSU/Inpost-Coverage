package domain

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type ExclusionArea struct {
	Category   string      `json:"category"`
	Bounds     BoundingBox `json:"bounds"`
	OuterRings [][]LatLng  `json:"outer_rings"`
	InnerRings [][]LatLng  `json:"inner_rings,omitempty"`
}
