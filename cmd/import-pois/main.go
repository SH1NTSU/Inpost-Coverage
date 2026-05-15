package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/adapter/client/overpass"
	pgrepo "github.com/marcelbudziszewski/paczkomat-predictor/internal/adapter/repository/postgres"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/config"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/logger"
	pgconn "github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/postgres"
)

func main() {
	province := flag.String("province", "", "Province name. Empty = iterate every province from the points table (smaller bboxes, friendlier to Overpass).")
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

	var provinces []string
	if *province != "" {
		provinces = []string{*province}
	} else {
		list, err := points.ListProvinces(ctx, 1)
		if err != nil {
			log.Error("list provinces", "err", err)
			return
		}
		for _, p := range list {
			provinces = append(provinces, p.Name)
		}
		log.Info("iterating provinces", "count", len(provinces))
	}

	var totalFetched int
	for i, name := range provinces {
		bb, err := points.BoundingBox(ctx, name)
		if err != nil {
			log.Error("bbox", "province", name, "err", err)
			continue
		}
		bb.MinLat -= 0.01
		bb.MinLng -= 0.015
		bb.MaxLat += 0.01
		bb.MaxLng += 0.015
		log.Info("fetching", "i", i+1, "of", len(provinces), "province", name,
			"bbox", fmt.Sprintf("%.3f,%.3f→%.3f,%.3f", bb.MinLat, bb.MinLng, bb.MaxLat, bb.MaxLng))

		items, err := op.FetchAnchorPOIs(ctx, bb)
		if err != nil {
			log.Error("overpass fetch", "province", name, "err", err)
			continue
		}
		if err := anchors.UpsertBatch(ctx, items); err != nil {
			log.Error("upsert", "province", name, "err", err)
			continue
		}
		totalFetched += len(items)
		log.Info("done", "province", name, "anchors", len(items))
	}

	total, _ := anchors.Count(ctx)
	log.Info("import finished", "total_in_db", total, "fetched_this_run", totalFetched)
}
