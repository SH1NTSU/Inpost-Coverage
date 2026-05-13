package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
)

type AnchorRepo struct {
	DB *pgxpool.Pool
}

func NewAnchorRepo(db *pgxpool.Pool) *AnchorRepo { return &AnchorRepo{DB: db} }

func (r *AnchorRepo) UpsertBatch(ctx context.Context, items []domain.AnchorPOI) error {
	if len(items) == 0 {
		return nil
	}
	const q = `
INSERT INTO anchor_pois (poi_type, brand, name, latitude, longitude, address, osm_id, raw_tags, fetched_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
ON CONFLICT (osm_id) DO UPDATE SET
    poi_type=EXCLUDED.poi_type,
    brand=EXCLUDED.brand,
    name=EXCLUDED.name,
    latitude=EXCLUDED.latitude,
    longitude=EXCLUDED.longitude,
    address=EXCLUDED.address,
    raw_tags=EXCLUDED.raw_tags,
    fetched_at=NOW()`
	batch := &pgx.Batch{}
	for _, p := range items {
		tags, _ := json.Marshal(map[string]string{
			"brand": p.Brand,
			"name":  p.Name,
			"type":  p.Type,
		})
		batch.Queue(q, p.Type, p.Brand, p.Name, p.Latitude, p.Longitude, p.Address, p.OSMID, string(tags))
	}
	br := r.DB.SendBatch(ctx, batch)
	defer br.Close()
	for range items {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *AnchorRepo) All(ctx context.Context, bbox domain.BoundingBox) ([]domain.AnchorPOI, error) {
	const q = `
SELECT id, poi_type, COALESCE(brand,''), COALESCE(name,''),
       latitude, longitude, COALESCE(address,''), osm_id, fetched_at
FROM anchor_pois
WHERE ($1::FLOAT8 = 0 AND $2::FLOAT8 = 0 AND $3::FLOAT8 = 0 AND $4::FLOAT8 = 0)
   OR (latitude BETWEEN $1 AND $3 AND longitude BETWEEN $2 AND $4)`
	rows, err := r.DB.Query(ctx, q, bbox.MinLat, bbox.MinLng, bbox.MaxLat, bbox.MaxLng)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.AnchorPOI
	for rows.Next() {
		var a domain.AnchorPOI
		if err := rows.Scan(
			&a.ID, &a.Type, &a.Brand, &a.Name,
			&a.Latitude, &a.Longitude, &a.Address, &a.OSMID, &a.FetchedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AnchorRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.DB.QueryRow(ctx, `SELECT COUNT(*) FROM anchor_pois`).Scan(&n)
	return n, err
}
