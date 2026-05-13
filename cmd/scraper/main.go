package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/adapter/client/inpost"
	pgrepo "github.com/marcelbudziszewski/paczkomat-predictor/internal/adapter/repository/postgres"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/config"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/logger"
	pgconn "github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/postgres"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/scheduler"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/usecase/scraping"
)

func main() {
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

	svc := &scraping.Service{
		Client:    inpost.New(cfg.External.InpostBaseURL, cfg.Scraper.RatePerSec),
		Points:    pgrepo.NewPointRepo(pool),
		Snapshots: pgrepo.NewSnapshotRepo(pool),
		Country:   cfg.Scraper.Country,
		City:      cfg.Scraper.City,
		PageSize:  cfg.Scraper.PageSize,
	}

	runScrape := func(ctx context.Context, phase string) {
		stats, err := svc.RunOnce(ctx)
		if err != nil {
			log.Error("scrape", "phase", phase, "err", err)
			return
		}
		log.Info("scrape done",
			"phase", phase,
			"points", stats.Points,
			"snapshots", stats.Snapshots,
			"duration", stats.FinishedAt.Sub(stats.StartedAt))
	}

	runScrape(ctx, "startup")

	sch := scheduler.New(log)
	if err := sch.Every(cfg.Scraper.Interval, "scrape", func(ctx context.Context) {
		runScrape(ctx, "scheduled")
	}); err != nil {
		log.Error("schedule", "err", err)
		return
	}
	sch.Start(ctx)
}
