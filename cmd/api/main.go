package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/adapter/client/overpass"
	pgrepo "github.com/marcelbudziszewski/paczkomat-predictor/internal/adapter/repository/postgres"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/cache"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/config"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/logger"
	pgconn "github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/postgres"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/infrastructure/scheduler"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/server"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/usecase/coverage"
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

	points := pgrepo.NewPointRepo(pool)
	competitors := pgrepo.NewCompetitorRepo(pool)
	anchors := pgrepo.NewAnchorRepo(pool)
	analytics := pgrepo.NewAnalyticsRepo(pool)
	coverageStore := pgrepo.NewCoverageCacheRepo(pool)
	terrain := overpass.New("")

	var c cache.Cache
	if cfg.Redis.URL != "" || cfg.Redis.Addr != "" {
		rc, err := cache.NewRedis(cache.RedisConfig{
			URL:      cfg.Redis.URL,
			Addr:     cfg.Redis.Addr,
			Username: cfg.Redis.Username,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}, cfg.Redis.TTL)
		if err != nil {
			log.Warn("redis unavailable, using in-memory cache", "err", err)
			c = cache.NewMemory(cfg.Redis.TTL)
		} else {
			target := cfg.Redis.Addr
			if cfg.Redis.URL != "" {
				target = "(REDIS_URL)"
			}
			log.Info("cache backend", "type", "redis", "addr", target, "ttl", cfg.Redis.TTL)
			c = rc
		}
	} else {
		log.Info("cache backend", "type", "memory", "ttl", cfg.Redis.TTL)
		c = cache.NewMemory(cfg.Redis.TTL)
	}

	cov := &coverage.Service{
		Points:      points,
		Competitors: competitors,
		Anchors:     anchors,
		Analytics:   analytics,
		Terrain:     terrain,
		Store:       coverageStore,
		Cache:       c,
	}

	runWarm := func(ctx context.Context, phase string) {
		stats, err := cov.WarmDefaults(
			ctx,
			cfg.Coverage.MinProvincePoints,
			cfg.Coverage.PrecomputeProvinceLimit,
			cfg.Coverage.DefaultCellMeters,
			cfg.Coverage.DefaultRecommendationsTop,
		)
		if err != nil {
			log.Error("coverage warm", "phase", phase, "err", err)
			return
		}
		log.Info("coverage warm done", "phase", phase, "provinces", stats.Provinces)
	}
	go runWarm(ctx, "startup")
	if cfg.Coverage.PrecomputeInterval > 0 {
		sch := scheduler.New(log)
		if err := sch.Every(cfg.Coverage.PrecomputeInterval, "coverage-warm", func(ctx context.Context) {
			runWarm(ctx, "scheduled")
		}); err != nil {
			log.Error("schedule", "job", "coverage-warm", "err", err)
		} else {
			go sch.Start(ctx)
		}
	}

	srv := &server.Server{
		Analytics:      analytics,
		Coverage:       cov,
		Competitors:    competitors,
		Points:         points,
		AllowedOrigins: cfg.HTTP.AllowedOrigins,
	}

	listenAddr := cfg.HTTP.Addr
	if cfg.HTTP.Port != "" {
		// Render/Heroku/Cloud Run convention: bind to 0.0.0.0:$PORT.
		listenAddr = ":" + cfg.HTTP.Port
	}
	httpSrv := &http.Server{
		Addr:         listenAddr,
		Handler:      srv.Router(),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("http listen", "addr", listenAddr)
		errCh <- httpSrv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http", "err", err)
		}
	case <-ctx.Done():
		log.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
	}
}
