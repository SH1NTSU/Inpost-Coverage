package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/adapter/client/inpost"
	pgrepo "github.com/marcelbudziszewski/paczkomat-predictor/internal/adapter/repository/postgres"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/config"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/logger"
	pgconn "github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/postgres"
)

func main() {
	country := flag.String("country", "PL", "ISO country code to import")
	city := flag.String("city", "", "Optional city filter; empty = whole country")
	pageSize := flag.Int("page-size", 100, "Items per page")
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

	client := inpost.New(cfg.External.InpostBaseURL, cfg.Scraper.RatePerSec)
	points := pgrepo.NewPointRepo(pool)

	start := time.Now()
	log.Info("starting import",
		"country", *country, "city", *city, "page_size", *pageSize)

	var total, upserted int
	err = client.ListByCity(ctx, *country, *city, *pageSize,
		func(page []domain.Point, _ []domain.AvailabilitySnapshot) error {
			for i := range page {
				if err := points.Upsert(ctx, &page[i]); err != nil {
					return err
				}
				upserted++
			}
			total += len(page)
			log.Info("progress", "page_total", len(page),
				"running_total", total, "elapsed", time.Since(start).Truncate(time.Second))
			return nil
		})
	if err != nil {
		log.Error("import", "err", err)
		return
	}
	log.Info("import done",
		"total", total, "upserted", upserted, "duration", time.Since(start).Truncate(time.Second))
}
