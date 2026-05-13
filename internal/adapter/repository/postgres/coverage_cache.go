package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
)

type CoverageCacheRepo struct {
	DB *pgxpool.Pool
}

func NewCoverageCacheRepo(db *pgxpool.Pool) *CoverageCacheRepo { return &CoverageCacheRepo{DB: db} }

const loadGridSQL = `
SELECT summary, cells
FROM coverage_grids
WHERE city = $1 AND cell_meters = $2 AND version = $3
`

func (r *CoverageCacheRepo) LoadGrid(ctx context.Context, city string, cellMeters int, version string) (*domain.CoverageGridSnapshot, error) {
	var summaryRaw []byte
	var cellsRaw []byte
	err := r.DB.QueryRow(ctx, loadGridSQL, city, cellMeters, version).Scan(&summaryRaw, &cellsRaw)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}

	var snap domain.CoverageGridSnapshot
	if err := json.Unmarshal(summaryRaw, &snap.Summary); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(cellsRaw, &snap.Cells); err != nil {
		return nil, err
	}
	return &snap, nil
}

const saveGridSQL = `
INSERT INTO coverage_grids (city, cell_meters, version, summary, cells, computed_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (city, cell_meters, version) DO UPDATE SET
    summary = EXCLUDED.summary,
    cells = EXCLUDED.cells,
    computed_at = NOW()
`

func (r *CoverageCacheRepo) SaveGrid(ctx context.Context, city string, cellMeters int, version string, snap domain.CoverageGridSnapshot) error {
	summaryRaw, err := json.Marshal(snap.Summary)
	if err != nil {
		return err
	}
	cellsRaw, err := json.Marshal(snap.Cells)
	if err != nil {
		return err
	}
	_, err = r.DB.Exec(ctx, saveGridSQL, city, cellMeters, version, summaryRaw, cellsRaw)
	return err
}

const loadRecommendationsSQL = `
SELECT payload
FROM coverage_recommendations
WHERE city = $1 AND limit_count = $2 AND version = $3
`

func (r *CoverageCacheRepo) LoadRecommendations(ctx context.Context, city string, limit int, version string) (*domain.CoverageRecommendations, error) {
	var raw []byte
	err := r.DB.QueryRow(ctx, loadRecommendationsSQL, city, limit, version).Scan(&raw)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	var out domain.CoverageRecommendations
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

const saveRecommendationsSQL = `
INSERT INTO coverage_recommendations (city, limit_count, version, payload, computed_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (city, limit_count, version) DO UPDATE SET
    payload = EXCLUDED.payload,
    computed_at = NOW()
`

func (r *CoverageCacheRepo) SaveRecommendations(ctx context.Context, city string, limit int, version string, recs domain.CoverageRecommendations) error {
	raw, err := json.Marshal(recs)
	if err != nil {
		return err
	}
	_, err = r.DB.Exec(ctx, saveRecommendationsSQL, city, limit, version, raw)
	return err
}
