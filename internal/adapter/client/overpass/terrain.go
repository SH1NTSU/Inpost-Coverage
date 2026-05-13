package overpass

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
)

const coordEpsilon = 1e-7

func (c *Client) FetchExclusionAreas(ctx context.Context, b domain.BoundingBox) ([]domain.ExclusionArea, error) {
	bbox := fmt.Sprintf("(%f,%f,%f,%f)", b.MinLat, b.MinLng, b.MaxLat, b.MaxLng)
	query := fmt.Sprintf(`[out:json][timeout:60];
(
  way["natural"="water"]%s;
  relation["natural"="water"]%s;
  way["water"]%s;
  relation["water"]%s;
  way["waterway"="riverbank"]%s;
  relation["waterway"="riverbank"]%s;
  way["natural"="wetland"]%s;
  relation["natural"="wetland"]%s;
  way["landuse"~"^(forest|farmland|meadow|orchard|vineyard)$"]%s;
  relation["landuse"~"^(forest|farmland|meadow|orchard|vineyard)$"]%s;
  way["natural"~"^(wood|scrub|heath|grassland)$"]%s;
  relation["natural"~"^(wood|scrub|heath|grassland)$"]%s;
);
out geom;`,
		bbox, bbox,
		bbox, bbox,
		bbox, bbox,
		bbox, bbox,
		bbox, bbox,
		bbox, bbox,
	)

	form := url.Values{"data": []string{query}}
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("overpass http %d", resp.StatusCode)
	}

	var body overpassResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	out := make([]domain.ExclusionArea, 0, len(body.Elements))
	for _, e := range body.Elements {
		area, ok := exclusionAreaFromElement(e)
		if !ok {
			continue
		}
		out = append(out, area)
	}
	return out, nil
}

func exclusionAreaFromElement(e overpassElement) (domain.ExclusionArea, bool) {
	category := classifyExclusionCategory(e.Tags)
	if category == "" {
		return domain.ExclusionArea{}, false
	}

	switch e.Type {
	case "way":
		ring := closeRing(coordsToPoints(e.Geometry))
		return buildExclusionArea(category, [][]domain.LatLng{ring}, nil)
	case "relation":
		var outerSegs [][]domain.LatLng
		var innerSegs [][]domain.LatLng
		for _, m := range e.Members {
			seg := coordsToPoints(m.Geometry)
			if len(seg) < 2 {
				continue
			}
			switch m.Role {
			case "outer":
				outerSegs = append(outerSegs, seg)
			case "inner":
				innerSegs = append(innerSegs, seg)
			}
		}
		return buildExclusionArea(category, assembleRings(outerSegs), assembleRings(innerSegs))
	default:
		return domain.ExclusionArea{}, false
	}
}

func buildExclusionArea(category string, outerRings, innerRings [][]domain.LatLng) (domain.ExclusionArea, bool) {
	if len(outerRings) == 0 {
		return domain.ExclusionArea{}, false
	}

	bounds := domain.BoundingBox{
		MinLat: math.Inf(1),
		MinLng: math.Inf(1),
		MaxLat: math.Inf(-1),
		MaxLng: math.Inf(-1),
	}
	validOuters := make([][]domain.LatLng, 0, len(outerRings))
	for _, ring := range outerRings {
		ring = closeRing(ring)
		if len(ring) < 4 {
			continue
		}
		validOuters = append(validOuters, ring)
		for _, p := range ring {
			if p.Lat < bounds.MinLat {
				bounds.MinLat = p.Lat
			}
			if p.Lat > bounds.MaxLat {
				bounds.MaxLat = p.Lat
			}
			if p.Lng < bounds.MinLng {
				bounds.MinLng = p.Lng
			}
			if p.Lng > bounds.MaxLng {
				bounds.MaxLng = p.Lng
			}
		}
	}
	if len(validOuters) == 0 {
		return domain.ExclusionArea{}, false
	}

	validInners := make([][]domain.LatLng, 0, len(innerRings))
	for _, ring := range innerRings {
		ring = closeRing(ring)
		if len(ring) < 4 {
			continue
		}
		validInners = append(validInners, ring)
	}

	return domain.ExclusionArea{
		Category:   category,
		Bounds:     bounds,
		OuterRings: validOuters,
		InnerRings: validInners,
	}, true
}

func classifyExclusionCategory(tags map[string]string) string {
	switch {
	case tags["natural"] == "water", tags["water"] != "", tags["waterway"] == "riverbank", tags["natural"] == "wetland":
		return "water"
	case tags["landuse"] == "forest", tags["natural"] == "wood":
		return "forest"
	case tags["landuse"] == "farmland", tags["landuse"] == "meadow", tags["landuse"] == "orchard", tags["landuse"] == "vineyard":
		return "field"
	case tags["natural"] == "scrub", tags["natural"] == "heath", tags["natural"] == "grassland":
		return "field"
	default:
		return ""
	}
}

func coordsToPoints(coords []overpassCoord) []domain.LatLng {
	out := make([]domain.LatLng, 0, len(coords))
	for _, c := range coords {
		out = append(out, domain.LatLng{Lat: c.Lat, Lng: c.Lon})
	}
	return out
}

func assembleRings(segments [][]domain.LatLng) [][]domain.LatLng {
	remaining := make([][]domain.LatLng, 0, len(segments))
	for _, seg := range segments {
		if len(seg) >= 2 {
			remaining = append(remaining, append([]domain.LatLng(nil), seg...))
		}
	}

	var rings [][]domain.LatLng
	for len(remaining) > 0 {
		ring := remaining[0]
		remaining = remaining[1:]

		for {
			if isClosed(ring) {
				break
			}
			matched := false
			for i, seg := range remaining {
				switch {
				case samePoint(ring[len(ring)-1], seg[0]):
					ring = append(ring, seg[1:]...)
				case samePoint(ring[len(ring)-1], seg[len(seg)-1]):
					reversed := reversePoints(seg)
					ring = append(ring, reversed[1:]...)
				case samePoint(ring[0], seg[len(seg)-1]):
					prefix := append([]domain.LatLng(nil), seg[:len(seg)-1]...)
					ring = append(prefix, ring...)
				case samePoint(ring[0], seg[0]):
					reversed := reversePoints(seg)
					prefix := append([]domain.LatLng(nil), reversed[:len(reversed)-1]...)
					ring = append(prefix, ring...)
				default:
					continue
				}
				remaining = append(remaining[:i], remaining[i+1:]...)
				matched = true
				break
			}
			if !matched {
				break
			}
		}

		ring = closeRing(ring)
		if len(ring) >= 4 && isClosed(ring) {
			rings = append(rings, ring)
		}
	}
	return rings
}

func closeRing(ring []domain.LatLng) []domain.LatLng {
	if len(ring) == 0 {
		return nil
	}
	if samePoint(ring[0], ring[len(ring)-1]) {
		return ring
	}
	return append(ring, ring[0])
}

func isClosed(ring []domain.LatLng) bool {
	return len(ring) >= 4 && samePoint(ring[0], ring[len(ring)-1])
}

func samePoint(a, b domain.LatLng) bool {
	return math.Abs(a.Lat-b.Lat) <= coordEpsilon && math.Abs(a.Lng-b.Lng) <= coordEpsilon
}

func reversePoints(in []domain.LatLng) []domain.LatLng {
	out := append([]domain.LatLng(nil), in...)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}
