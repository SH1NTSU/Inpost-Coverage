package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
)

type CompetitorRepo struct {
	DB *pgxpool.Pool
}

func NewCompetitorRepo(db *pgxpool.Pool) *CompetitorRepo { return &CompetitorRepo{DB: db} }

type competitorRow struct {
	domain.CompetitorPoint
	RawTags map[string]string
}

func (r *CompetitorRepo) UpsertBatch(ctx context.Context, items []domain.CompetitorPoint) error {
	if len(items) == 0 {
		return nil
	}
	const q = `
INSERT INTO competitor_points (network, name, latitude, longitude, address, osm_id, raw_tags, fetched_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
ON CONFLICT (osm_id) DO UPDATE SET
    network=EXCLUDED.network,
    name=EXCLUDED.name,
    latitude=EXCLUDED.latitude,
    longitude=EXCLUDED.longitude,
    address=EXCLUDED.address,
    raw_tags=EXCLUDED.raw_tags,
    fetched_at=NOW()`
	batch := &pgx.Batch{}
	for _, c := range items {

		tags, _ := json.Marshal(map[string]string{"network": c.Network, "name": c.Name})
		batch.Queue(q, c.Network, c.Name, c.Latitude, c.Longitude, c.Address, c.OSMID, string(tags))
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

func (r *CompetitorRepo) AllForCoverage(ctx context.Context, bbox domain.BoundingBox) ([]domain.CompetitorPoint, error) {
	const q = `
SELECT id, network, COALESCE(name,''), latitude, longitude, COALESCE(address,''), osm_id, fetched_at
FROM competitor_points
WHERE ($1::FLOAT8 = 0 AND $2::FLOAT8 = 0 AND $3::FLOAT8 = 0 AND $4::FLOAT8 = 0)
   OR (latitude BETWEEN $1 AND $3 AND longitude BETWEEN $2 AND $4)`
	rows, err := r.DB.Query(ctx, q, bbox.MinLat, bbox.MinLng, bbox.MaxLat, bbox.MaxLng)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.CompetitorPoint
	for rows.Next() {
		var c domain.CompetitorPoint
		if err := rows.Scan(
			&c.ID, &c.Network, &c.Name, &c.Latitude, &c.Longitude, &c.Address, &c.OSMID, &c.FetchedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CompetitorRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.DB.QueryRow(ctx, `SELECT COUNT(*) FROM competitor_points`).Scan(&n)
	return n, err
}
