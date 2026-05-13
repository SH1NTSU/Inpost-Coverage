package overpass

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client

	UserAgent string
}

func New(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://overpass-api.de/api/interpreter"
	}
	return &Client{
		BaseURL:   baseURL,
		HTTP:      &http.Client{Timeout: 60 * time.Second},
		UserAgent: "paczkomat-reliability/0.1 (educational; github.com/marcelbudziszewski/paczkomat-predictor)",
	}
}

func (c *Client) FetchParcelLockers(ctx context.Context, b domain.BoundingBox) ([]domain.CompetitorPoint, error) {

	query := fmt.Sprintf(`[out:json][timeout:60];
(
  node["amenity"="parcel_locker"](%f,%f,%f,%f);
  node["amenity"="vending_machine"]["vending"="parcel_pickup"](%f,%f,%f,%f);
  node["amenity"="post_office"](%f,%f,%f,%f);
);
out body;`,
		b.MinLat, b.MinLng, b.MaxLat, b.MaxLng,
		b.MinLat, b.MinLng, b.MaxLat, b.MaxLng,
		b.MinLat, b.MinLng, b.MaxLat, b.MaxLng,
	)

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL,
		strings.NewReader("data="+query))
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

	out := make([]domain.CompetitorPoint, 0, len(body.Elements))
	for _, e := range body.Elements {
		network := classifyBrand(e.Tags)

		if network == "InPost" || network == "" {
			continue
		}
		out = append(out, domain.CompetitorPoint{
			Network:   network,
			Name:      e.Tags["name"],
			Latitude:  e.Lat,
			Longitude: e.Lon,
			Address:   buildAddress(e.Tags),
			OSMID:     e.ID,
		})
	}
	return out, nil
}

func (c *Client) FetchAnchorPOIs(ctx context.Context, b domain.BoundingBox) ([]domain.AnchorPOI, error) {

	query := fmt.Sprintf(`[out:json][timeout:60];
(
  node["shop"="convenience"](%f,%f,%f,%f);
  node["shop"="supermarket"](%f,%f,%f,%f);
  node["shop"="mall"](%f,%f,%f,%f);
  node["amenity"="fuel"](%f,%f,%f,%f);
  node["amenity"="marketplace"](%f,%f,%f,%f);
  node["amenity"="university"](%f,%f,%f,%f);
  node["amenity"="college"](%f,%f,%f,%f);
  node["public_transport"="station"](%f,%f,%f,%f);
  node["railway"="station"](%f,%f,%f,%f);
);
out body;`,
		b.MinLat, b.MinLng, b.MaxLat, b.MaxLng,
		b.MinLat, b.MinLng, b.MaxLat, b.MaxLng,
		b.MinLat, b.MinLng, b.MaxLat, b.MaxLng,
		b.MinLat, b.MinLng, b.MaxLat, b.MaxLng,
		b.MinLat, b.MinLng, b.MaxLat, b.MaxLng,
		b.MinLat, b.MinLng, b.MaxLat, b.MaxLng,
		b.MinLat, b.MinLng, b.MaxLat, b.MaxLng,
		b.MinLat, b.MinLng, b.MaxLat, b.MaxLng,
		b.MinLat, b.MinLng, b.MaxLat, b.MaxLng,
	)

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL,
		strings.NewReader("data="+query))
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

	out := make([]domain.AnchorPOI, 0, len(body.Elements))
	for _, e := range body.Elements {
		typ := classifyAnchorType(e.Tags)
		if typ == "" {
			continue
		}
		out = append(out, domain.AnchorPOI{
			Type:      typ,
			Brand:     pickBrand(e.Tags),
			Name:      e.Tags["name"],
			Latitude:  e.Lat,
			Longitude: e.Lon,
			Address:   buildAddress(e.Tags),
			OSMID:     e.ID,
		})
	}
	return out, nil
}

func classifyAnchorType(t map[string]string) string {
	switch {
	case t["shop"] == "convenience":
		return "convenience"
	case t["shop"] == "supermarket":
		return "supermarket"
	case t["shop"] == "mall":
		return "mall"
	case t["amenity"] == "fuel":
		return "fuel"
	case t["amenity"] == "marketplace":
		return "marketplace"
	case t["amenity"] == "university", t["amenity"] == "college":
		return "university"
	case t["public_transport"] == "station", t["railway"] == "station":
		return "transit"
	}
	return ""
}

func pickBrand(t map[string]string) string {
	for _, k := range []string{"brand", "operator", "name"} {
		if v := strings.TrimSpace(t[k]); v != "" {
			return v
		}
	}
	return ""
}

type overpassResponse struct {
	Elements []overpassElement `json:"elements"`
}

type overpassElement struct {
	Type     string            `json:"type"`
	ID       int64             `json:"id"`
	Lat      float64           `json:"lat"`
	Lon      float64           `json:"lon"`
	Tags     map[string]string `json:"tags"`
	Geometry []overpassCoord   `json:"geometry"`
	Members  []overpassMember  `json:"members"`
}

type overpassCoord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type overpassMember struct {
	Type     string          `json:"type"`
	Ref      int64           `json:"ref"`
	Role     string          `json:"role"`
	Geometry []overpassCoord `json:"geometry"`
}

func classifyBrand(t map[string]string) string {
	probe := strings.ToLower(strings.Join([]string{
		t["brand"], t["operator"], t["name"],
	}, " "))

	switch {
	case strings.Contains(probe, "inpost") || strings.Contains(probe, "paczkomat"):
		return "InPost"
	case strings.Contains(probe, "allegro one") || strings.Contains(probe, "allegrobox") || strings.Contains(probe, "one box"):
		return "AllegroOne"
	case strings.Contains(probe, "dhl"):
		return "DHL"
	case strings.Contains(probe, "orlen"):
		return "OrlenPaczka"
	case strings.Contains(probe, "poczta polska") || strings.Contains(probe, "pocztex"):
		return "PocztaPolska"
	case strings.Contains(probe, "dpd"):
		return "DPD"
	case strings.Contains(probe, "gls"):
		return "GLS"
	case strings.Contains(probe, "fedex"):
		return "FedEx"
	case strings.Contains(probe, "ups"):
		return "UPS"
	}

	return ""
}

func buildAddress(t map[string]string) string {
	parts := []string{}
	if v := t["addr:street"]; v != "" {
		if hn := t["addr:housenumber"]; hn != "" {
			parts = append(parts, v+" "+hn)
		} else {
			parts = append(parts, v)
		}
	}
	if v := t["addr:city"]; v != "" {
		parts = append(parts, v)
	}
	return strings.Join(parts, ", ")
}
