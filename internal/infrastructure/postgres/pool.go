package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/config"
)

func Open(ctx context.Context, cfg config.DB) (*pgxpool.Pool, error) {
	pcfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, err
	}
	pcfg.MaxConns = cfg.MaxConns
	pcfg.MaxConnIdleTime = 5 * time.Minute

	// Neon's PgBouncer pooler strips `search_path` startup parameters, and
	// ALTER ROLE / ALTER DATABASE search_path settings don't propagate through
	// the pooler either. The result is connections land with an empty
	// search_path and our unqualified table references (FROM points, …) fail.
	// Setting it explicitly on every new pooled connection is the documented
	// workaround and is a no-op against vanilla Postgres.
	pcfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET search_path TO public")
		return err
	}

	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
