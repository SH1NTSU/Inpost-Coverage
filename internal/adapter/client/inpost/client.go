package inpost

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
	"github.com/marcelbudziszewski/paczkomat-predictor/pkg/httpx"
)

type Client struct {
	HTTP *httpx.Client
}

func New(baseURL string, rps int) *Client {
	return &Client{HTTP: httpx.New(baseURL, 15*time.Second, httpx.NewTokenBucket(rps))}
}

func (c *Client) ListByCity(ctx context.Context, country, city string, pageSize int,
	yield func(page []domain.Point, snapshots []domain.AvailabilitySnapshot) error) error {
	if pageSize == 0 {
		pageSize = 100
	}
	now := time.Now().UTC()
	for page := 1; ; page++ {
		q := url.Values{}

		if city != "" {
			q.Set("city", city)
		}
		if country != "" {
			q.Set("country", country)
		}
		q.Set("status", "Operating")
		q.Set("per_page", fmt.Sprintf("%d", pageSize))
		q.Set("page", fmt.Sprintf("%d", page))

		var resp pointsResponse
		if err := c.HTTP.Do(ctx, "GET", "/v1/points?"+q.Encode(), nil, &resp); err != nil {
			return err
		}
		if len(resp.Items) == 0 {
			return nil
		}

		points := make([]domain.Point, 0, len(resp.Items))
		snaps := make([]domain.AvailabilitySnapshot, 0, len(resp.Items))
		for _, item := range resp.Items {
			p, s := item.toDomain(now)
			points = append(points, p)
			snaps = append(snaps, s)
		}
		if err := yield(points, snaps); err != nil {
			return err
		}
		if page >= resp.TotalPages {
			return nil
		}
	}
}

type pointsResponse struct {
	Items      []apiPoint `json:"items"`
	TotalPages int        `json:"total_pages"`
}

type apiPoint struct {
	Name         string `json:"name"`
	Status       string `json:"status"`
	LocationType string `json:"location_type"`
	Location247  bool   `json:"location_247"`
	IsNext       bool   `json:"is_next"`
	PhysicalType string `json:"physical_type"`
	ImageURL     string `json:"image_url"`
	Location     struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
	AddressDetails struct {
		City           string `json:"city"`
		Province       string `json:"province"`
		PostCode       string `json:"post_code"`
		Street         string `json:"street"`
		BuildingNumber string `json:"building_number"`
	} `json:"address_details"`
	Country string `json:"country"`
}

func (a apiPoint) toDomain(at time.Time) (domain.Point, domain.AvailabilitySnapshot) {
	p := domain.Point{
		InpostID:     a.Name,
		Country:      a.Country,
		Status:       domain.PointStatus(a.Status),
		Latitude:     a.Location.Latitude,
		Longitude:    a.Location.Longitude,
		City:         a.AddressDetails.City,
		Province:     a.AddressDetails.Province,
		PostCode:     a.AddressDetails.PostCode,
		Street:       a.AddressDetails.Street,
		BuildingNo:   a.AddressDetails.BuildingNumber,
		LocationType: a.LocationType,
		Location247:  a.Location247,
		IsNext:       a.IsNext,
		PhysicalType: a.PhysicalType,
		ImageURL:     a.ImageURL,
	}
	s := domain.AvailabilitySnapshot{
		CapturedAt: at,
		Status:     p.Status,
	}
	return p, s
}
