package domain

import "time"

type Stats struct {
	TotalLockers     int        `json:"total_lockers"`
	Operating        int        `json:"operating"`
	Disabled         int        `json:"disabled"`
	SnapshotsTotal   int64      `json:"snapshots_total"`
	ScraperLastRunAt *time.Time `json:"scraper_last_run_at"`
}

type LockerSummary struct {
	ID            int64       `json:"id"`
	InpostID      string      `json:"inpost_id"`
	City          string      `json:"city"`
	Street        string      `json:"street"`
	BuildingNo    string      `json:"building_no"`
	Latitude      float64     `json:"latitude"`
	Longitude     float64     `json:"longitude"`
	ImageURL      string      `json:"image_url"`
	Location247   bool        `json:"location_247"`
	IsNext        bool        `json:"is_next"`
	CurrentStatus PointStatus `json:"current_status"`
	LastChangeAt  *time.Time  `json:"last_change_at"`
}

type LockerDetail struct {
	LockerSummary
	Country                string     `json:"country"`
	Province               string     `json:"province"`
	PostCode               string     `json:"post_code"`
	LocationType           string     `json:"location_type"`
	PhysicalType           string     `json:"physical_type"`
	CurrentOutageStartedAt *time.Time `json:"current_outage_started_at"`
	Uptime24hPct           float64    `json:"uptime_24h_pct"`
	Uptime7dPct            float64    `json:"uptime_7d_pct"`
	UptimeAllPct           float64    `json:"uptime_all_pct"`
	StateChanges24h        int        `json:"state_changes_24h"`
}

type StateChange struct {
	ChangedAt  time.Time   `json:"changed_at"`
	FromStatus PointStatus `json:"from_status"`
	ToStatus   PointStatus `json:"to_status"`
}

type OngoingOutage struct {
	Locker          LockerSummary `json:"locker"`
	StartedAt       time.Time     `json:"started_at"`
	DurationSeconds int64         `json:"duration_seconds"`
}

type WorstOffender struct {
	Locker          LockerSummary `json:"locker"`
	OutageEvents    int           `json:"outage_events"`
	DowntimeSeconds int64         `json:"downtime_seconds"`
	UptimePct       float64       `json:"uptime_pct"`
}

type DailyOutages struct {
	Day          time.Time `json:"day"`
	OutageEvents int       `json:"outage_events"`
}

type UptimeBucket struct {
	Label   string `json:"label"`
	Lockers int    `json:"lockers"`
}

type NetworkStats struct {
	WindowDays               int     `json:"window_days"`
	OutageEvents             int     `json:"outage_events"`
	AvgOutageDurationSeconds int64   `json:"avg_outage_duration_seconds"`
	NetworkAvailabilityPct   float64 `json:"network_availability_pct"`
}
