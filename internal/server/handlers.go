package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

type errorBody struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorBody{Error: "not found"})
	case errors.Is(err, domain.ErrInvalidArgument):
		writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, errorBody{Error: "internal error"})
	}
}

func jsonArray[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.Analytics.Stats(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) handleListLockers(w http.ResponseWriter, r *http.Request) {
	var filter *domain.PointStatus
	city := r.URL.Query().Get("city")
	if raw := r.URL.Query().Get("status"); raw != "" {
		switch domain.PointStatus(raw) {
		case domain.StatusOperating, domain.StatusDisabled:
			st := domain.PointStatus(raw)
			filter = &st
		default:
			writeError(w, domain.ErrInvalidArgument)
			return
		}
	}
	lockers, err := s.Analytics.ListLockers(r.Context(), filter, city)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, jsonArray(lockers))
}

func (s *Server) handleGetLocker(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, domain.ErrInvalidArgument)
		return
	}
	locker, err := s.Analytics.GetLocker(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, locker)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, domain.ErrInvalidArgument)
		return
	}
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	changes, err := s.Analytics.History(r.Context(), id, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, jsonArray(changes))
}

func (s *Server) handleCurrentOutages(w http.ResponseWriter, r *http.Request) {
	outages, err := s.Analytics.CurrentOutages(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, jsonArray(outages))
}

func intQuery(r *http.Request, name string, def int) int {
	if raw := r.URL.Query().Get(name); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func (s *Server) handleWorstOffenders(w http.ResponseWriter, r *http.Request) {
	days := intQuery(r, "days", 7)
	limit := intQuery(r, "limit", 10)
	out, err := s.Analytics.WorstOffenders(r.Context(), days, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, jsonArray(out))
}

func (s *Server) handleOutagesTimeline(w http.ResponseWriter, r *http.Request) {
	days := intQuery(r, "days", 14)
	out, err := s.Analytics.OutagesTimeline(r.Context(), days)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, jsonArray(out))
}

func (s *Server) handleUptimeDistribution(w http.ResponseWriter, r *http.Request) {
	days := intQuery(r, "days", 7)
	out, err := s.Analytics.UptimeDistribution(r.Context(), days)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, jsonArray(out))
}

func (s *Server) handleCoverageGrid(w http.ResponseWriter, r *http.Request) {
	cellM := intQuery(r, "cell_m", 300)
	city := r.URL.Query().Get("city")
	snap, err := s.Coverage.Grid(r.Context(), city, cellM)
	if err != nil {
		writeError(w, err)
		return
	}
	if snap.Cells == nil {
		snap.Cells = []domain.GridCell{}
	}
	writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleCoverageSummary(w http.ResponseWriter, r *http.Request) {
	cellM := intQuery(r, "cell_m", 300)
	city := r.URL.Query().Get("city")
	snap, err := s.Coverage.Grid(r.Context(), city, cellM)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snap.Summary)
}

func (s *Server) handleCoverageGridCells(w http.ResponseWriter, r *http.Request) {
	cellM := intQuery(r, "cell_m", 300)
	city := r.URL.Query().Get("city")
	snap, err := s.Coverage.Grid(r.Context(), city, cellM)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, jsonArray(snap.Cells))
}

func (s *Server) handleCoverageRecommendations(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	limit := intQuery(r, "limit", 10)
	out, err := s.Coverage.Recommendations(r.Context(), city, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	if out.NewPoints == nil {
		out.NewPoints = []domain.GapSuggestion{}
	}
	if out.Upgrades == nil {
		out.Upgrades = []domain.UpgradeCandidate{}
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleCoverageCompetitors(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	bb, err := s.Points.BoundingBox(r.Context(), city)
	if err != nil {
		writeError(w, err)
		return
	}

	if !bb.IsZero() {
		bb.MinLat -= 0.01
		bb.MinLng -= 0.015
		bb.MaxLat += 0.01
		bb.MaxLng += 0.015
	}
	out, err := s.Competitors.AllForCoverage(r.Context(), bb)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, jsonArray(out))
}

func (s *Server) handleCities(w http.ResponseWriter, r *http.Request) {
	min := intQuery(r, "min_points", 20)
	out, err := s.Points.ListCities(r.Context(), min)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, jsonArray(out))
}

func (s *Server) handleNetworkStats(w http.ResponseWriter, r *http.Request) {
	days := intQuery(r, "days", 7)
	out, err := s.Analytics.NetworkStats(r.Context(), days)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
