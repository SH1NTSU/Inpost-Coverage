package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
)

type SnapshotRepo struct {
	DB *pgxpool.Pool
}

func NewSnapshotRepo(db *pgxpool.Pool) *SnapshotRepo { return &SnapshotRepo{DB: db} }

func (r *SnapshotRepo) InsertBatch(ctx context.Context, items []domain.AvailabilitySnapshot) error {
	if len(items) == 0 {
		return nil
	}

	rows := make([][]any, len(items))
	for i, s := range items {
		rows[i] = []any{s.PointID, s.CapturedAt, s.Status}
	}

	_, err := r.DB.CopyFrom(ctx,
		pgx.Identifier{"availability_snapshots"},
		[]string{"point_id", "captured_at", "status"},
		pgx.CopyFromRows(rows),
	)
	return err
}
