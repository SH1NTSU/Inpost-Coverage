package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
	"github.com/marcelbudziszewski/paczkomat-predictor/internal/usecase/coverage"
)

type Server struct {
	Analytics      domain.AnalyticsRepository
	Coverage       *coverage.Service
	Competitors    domain.CompetitorRepository
	Points         domain.PointRepository
	AllowedOrigins []string
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Logger)
	r.Use(chimw.Timeout(20 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.AllowedOrigins,
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/stats", s.handleStats)
		r.Get("/lockers", s.handleListLockers)
		r.Get("/lockers/{id}", s.handleGetLocker)
		r.Get("/lockers/{id}/history", s.handleHistory)
		r.Get("/outages/current", s.handleCurrentOutages)

		r.Get("/stats/network", s.handleNetworkStats)
		r.Get("/stats/worst-offenders", s.handleWorstOffenders)
		r.Get("/stats/outages-timeline", s.handleOutagesTimeline)
		r.Get("/stats/uptime-distribution", s.handleUptimeDistribution)

		r.Get("/coverage/summary", s.handleCoverageSummary)
		r.Get("/coverage/grid-cells", s.handleCoverageGridCells)
		r.Get("/coverage/grid", s.handleCoverageGrid)
		r.Get("/coverage/recommendations", s.handleCoverageRecommendations)
		r.Get("/coverage/competitors", s.handleCoverageCompetitors)

		r.Get("/provinces", s.handleProvinces)
	})

	return r
}
