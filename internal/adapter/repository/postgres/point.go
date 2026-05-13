package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
)

type PointRepo struct {
	DB *pgxpool.Pool
}

func NewPointRepo(db *pgxpool.Pool) *PointRepo { return &PointRepo{DB: db} }

const upsertPointSQL = `
INSERT INTO points (
    inpost_id, country, status, latitude, longitude, city, province,
    post_code, street, building_no, location_type, is_next, location_247,
    physical_type, image_url, updated_at
) VALUES (
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,NOW()
)
ON CONFLICT (inpost_id) DO UPDATE SET
    country=EXCLUDED.country,
    status=EXCLUDED.status,
    latitude=EXCLUDED.latitude,
    longitude=EXCLUDED.longitude,
    city=EXCLUDED.city,
    province=EXCLUDED.province,
    post_code=EXCLUDED.post_code,
    street=EXCLUDED.street,
    building_no=EXCLUDED.building_no,
    location_type=EXCLUDED.location_type,
    is_next=EXCLUDED.is_next,
    location_247=EXCLUDED.location_247,
    physical_type=EXCLUDED.physical_type,
    image_url=EXCLUDED.image_url,
    updated_at=NOW()
RETURNING id;
`

func (r *PointRepo) Upsert(ctx context.Context, p *domain.Point) error {
	return r.DB.QueryRow(ctx, upsertPointSQL,
		p.InpostID, p.Country, p.Status, p.Latitude, p.Longitude, p.City, p.Province,
		p.PostCode, p.Street, p.BuildingNo, p.LocationType, p.IsNext, p.Location247,
		p.PhysicalType, p.ImageURL,
	).Scan(&p.ID)
}

func (r *PointRepo) AllForCoverage(ctx context.Context, city string) ([]domain.CoveragePoint, error) {
	const q = `
SELECT id, latitude, longitude, COALESCE(is_next, false)
FROM points
WHERE ($1::TEXT = '' OR city = $1)`
	rows, err := r.DB.Query(ctx, q, city)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.CoveragePoint
	for rows.Next() {
		var p domain.CoveragePoint
		if err := rows.Scan(&p.ID, &p.Latitude, &p.Longitude, &p.IsNext); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PointRepo) BoundingBox(ctx context.Context, city string) (domain.BoundingBox, error) {
	const q = `
SELECT MIN(latitude), MIN(longitude), MAX(latitude), MAX(longitude)
FROM points
WHERE ($1::TEXT = '' OR city = $1)`
	var b domain.BoundingBox
	err := r.DB.QueryRow(ctx, q, city).Scan(&b.MinLat, &b.MinLng, &b.MaxLat, &b.MaxLng)
	return b, err
}

func (r *PointRepo) ListCities(ctx context.Context, minPoints int) ([]domain.CityInfo, error) {
	if minPoints < 1 {
		minPoints = 1
	}
	const q = `
SELECT
    city,
    COALESCE(MAX(province), '')      AS province,
    COUNT(*)::INT                     AS point_count,
    MIN(latitude)                    AS min_lat,
    MIN(longitude)                   AS min_lng,
    MAX(latitude)                    AS max_lat,
    MAX(longitude)                   AS max_lng,
    (MIN(latitude) + MAX(latitude)) / 2  AS center_lat,
    (MIN(longitude) + MAX(longitude)) / 2 AS center_lng
FROM points
WHERE city IS NOT NULL AND city <> ''
GROUP BY city
HAVING COUNT(*) >= $1
ORDER BY COUNT(*) DESC, city ASC`
	rows, err := r.DB.Query(ctx, q, minPoints)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.CityInfo
	for rows.Next() {
		var c domain.CityInfo
		if err := rows.Scan(
			&c.Name, &c.Province, &c.PointCount,
			&c.MinLat, &c.MinLng, &c.MaxLat, &c.MaxLng,
			&c.CenterLat, &c.CenterLng,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
