package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/adapter/client/overpass"
	pgrepo "github.com/marcelbudziszewski/paczkomat-predictor/internal/adapter/repository/postgres"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/config"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/logger"
	pgconn "github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/postgres"
)

func main() {
	city := flag.String("city", "", "Optional city to constrain the bbox; empty = whole point set")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.New(cfg.Log.Level, cfg.Log.Format)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgconn.Open(ctx, cfg.DB)
	if err != nil {
		log.Error("db open", "err", err)
		return
	}
	defer pool.Close()

	points := pgrepo.NewPointRepo(pool)
	anchors := pgrepo.NewAnchorRepo(pool)
	op := overpass.New("https://overpass-api.de/api/interpreter")

	bb, err := points.BoundingBox(ctx, *city)
	if err != nil {
		log.Error("bbox", "err", err)
		return
	}
	bb.MinLat -= 0.01
	bb.MinLng -= 0.015
	bb.MaxLat += 0.01
	bb.MaxLng += 0.015
	log.Info("bounding box", "city", *city,
		"min_lat", bb.MinLat, "min_lng", bb.MinLng,
		"max_lat", bb.MaxLat, "max_lng", bb.MaxLng)

	log.Info("fetching anchor POIs from Overpass…")
	items, err := op.FetchAnchorPOIs(ctx, bb)
	if err != nil {
		log.Error("overpass fetch", "err", err)
		return
	}
	log.Info("fetched", "anchors", len(items))

	if err := anchors.UpsertBatch(ctx, items); err != nil {
		log.Error("upsert", "err", err)
		return
	}
	total, _ := anchors.Count(ctx)
	log.Info("import done", "total_in_db", total)
}
