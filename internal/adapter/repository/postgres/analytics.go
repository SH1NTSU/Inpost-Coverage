package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/marcelbudziszewski/paczkomat-predictor/internal/domain"
)

type AnalyticsRepo struct {
	DB *pgxpool.Pool
}

func NewAnalyticsRepo(db *pgxpool.Pool) *AnalyticsRepo { return &AnalyticsRepo{DB: db} }

const statsSQL = `
WITH latest AS (
    SELECT DISTINCT ON (point_id) point_id, status
    FROM availability_snapshots
    ORDER BY point_id, captured_at DESC
)
SELECT
    (SELECT COUNT(*) FROM points)::INT                      AS total_lockers,
    COALESCE(COUNT(*) FILTER (WHERE l.status = 'Operating'), 0)::INT AS operating,
    COALESCE(COUNT(*) FILTER (WHERE l.status = 'Disabled'),  0)::INT AS disabled,
    (SELECT COUNT(*) FROM availability_snapshots)::BIGINT   AS snapshots_total,
    (SELECT MAX(captured_at) FROM availability_snapshots)   AS scraper_last_run_at
FROM latest l;
`

func (r *AnalyticsRepo) Stats(ctx context.Context) (domain.Stats, error) {
	var s domain.Stats
	err := r.DB.QueryRow(ctx, statsSQL).Scan(
		&s.TotalLockers, &s.Operating, &s.Disabled,
		&s.SnapshotsTotal, &s.ScraperLastRunAt,
	)
	return s, err
}

const listLockersSQL = `
SELECT
    p.id, p.inpost_id, p.city, p.street, p.building_no,
    p.latitude, p.longitude, p.image_url, p.location_247, p.is_next,
    COALESCE(latest.status, p.status) AS current_status,
    NULL::TIMESTAMPTZ AS last_change_at
FROM points p
LEFT JOIN LATERAL (
    SELECT status
    FROM availability_snapshots s
    WHERE s.point_id = p.id
    ORDER BY s.captured_at DESC
    LIMIT 1
) latest ON TRUE
WHERE ($1::TEXT IS NULL OR COALESCE(latest.status, p.status) = $1)
  AND ($2::TEXT = '' OR p.city = $2)
ORDER BY p.id;
`

func (r *AnalyticsRepo) ListLockers(ctx context.Context, statusFilter *domain.PointStatus, city string) ([]domain.LockerSummary, error) {
	var arg any
	if statusFilter != nil {
		arg = string(*statusFilter)
	}
	rows, err := r.DB.Query(ctx, listLockersSQL, arg, city)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.LockerSummary
	for rows.Next() {
		var l domain.LockerSummary
		if err := rows.Scan(
			&l.ID, &l.InpostID, &l.City, &l.Street, &l.BuildingNo,
			&l.Latitude, &l.Longitude, &l.ImageURL, &l.Location247, &l.IsNext,
			&l.CurrentStatus, &l.LastChangeAt,
		); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

const getLockerSQL = `
WITH snaps AS (
    SELECT captured_at, status
    FROM availability_snapshots
    WHERE point_id = $1
),
latest AS (
    SELECT status, captured_at
    FROM snaps
    ORDER BY captured_at DESC
    LIMIT 1
),
runs AS (
    SELECT
        captured_at,
        status,
        LAG(status) OVER (ORDER BY captured_at) AS prev_status
    FROM snaps
),
last_change AS (
    SELECT captured_at
    FROM runs
    WHERE prev_status IS NOT NULL AND prev_status IS DISTINCT FROM status
    ORDER BY captured_at DESC
    LIMIT 1
),
last_disabled_start AS (
    SELECT captured_at
    FROM runs
    WHERE prev_status = 'Operating' AND status = 'Disabled'
    ORDER BY captured_at DESC
    LIMIT 1
),
changes_24h AS (
    SELECT COUNT(*) AS n
    FROM runs
    WHERE prev_status IS NOT NULL
      AND prev_status IS DISTINCT FROM status
      AND captured_at >= NOW() - INTERVAL '24 hours'
),
uptime AS (
    SELECT
        COALESCE(AVG((status = 'Operating')::INT::FLOAT) FILTER (WHERE captured_at >= NOW() - INTERVAL '24 hours'), 1) AS u24,
        COALESCE(AVG((status = 'Operating')::INT::FLOAT) FILTER (WHERE captured_at >= NOW() - INTERVAL '7 days'),  1) AS u7d,
        COALESCE(AVG((status = 'Operating')::INT::FLOAT),                                                          1) AS uall
    FROM snaps
)
SELECT
    p.id, p.inpost_id, p.city, p.street, p.building_no,
    p.latitude, p.longitude, p.image_url, p.location_247, p.is_next,
    p.country, p.province, p.post_code, p.location_type, p.physical_type,
    COALESCE(latest.status, p.status)                              AS current_status,
    (SELECT captured_at FROM last_change)                          AS last_change_at,
    CASE WHEN COALESCE(latest.status, p.status) = 'Disabled'
         THEN (SELECT captured_at FROM last_disabled_start) END    AS current_outage_started_at,
    uptime.u24, uptime.u7d, uptime.uall,
    COALESCE((SELECT n FROM changes_24h), 0)::INT                  AS state_changes_24h
FROM points p
LEFT JOIN latest ON TRUE
CROSS JOIN uptime
WHERE p.id = $1;
`

func (r *AnalyticsRepo) GetLocker(ctx context.Context, id int64) (*domain.LockerDetail, error) {
	var d domain.LockerDetail
	err := r.DB.QueryRow(ctx, getLockerSQL, id).Scan(
		&d.ID, &d.InpostID, &d.City, &d.Street, &d.BuildingNo,
		&d.Latitude, &d.Longitude, &d.ImageURL, &d.Location247, &d.IsNext,
		&d.Country, &d.Province, &d.PostCode, &d.LocationType, &d.PhysicalType,
		&d.CurrentStatus, &d.LastChangeAt, &d.CurrentOutageStartedAt,
		&d.Uptime24hPct, &d.Uptime7dPct, &d.UptimeAllPct,
		&d.StateChanges24h,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

const historySQL = `
WITH runs AS (
    SELECT
        captured_at,
        status,
        LAG(status) OVER (ORDER BY captured_at) AS prev_status
    FROM availability_snapshots
    WHERE point_id = $1
)
SELECT captured_at, prev_status, status
FROM runs
WHERE prev_status IS NOT NULL AND prev_status IS DISTINCT FROM status
ORDER BY captured_at DESC
LIMIT $2;
`

func (r *AnalyticsRepo) History(ctx context.Context, id int64, limit int) ([]domain.StateChange, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	rows, err := r.DB.Query(ctx, historySQL, id, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.StateChange
	for rows.Next() {
		var c domain.StateChange
		if err := rows.Scan(&c.ChangedAt, &c.FromStatus, &c.ToStatus); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

const outagesSQL = `
WITH latest AS (
    SELECT DISTINCT ON (point_id) point_id, status, captured_at
    FROM availability_snapshots
    ORDER BY point_id, captured_at DESC
),
disabled_now AS (
    SELECT point_id, captured_at AS latest_at
    FROM latest
    WHERE status = 'Disabled'
),
runs AS (
    SELECT
        s.point_id,
        s.captured_at,
        s.status,
        LAG(s.status) OVER (PARTITION BY s.point_id ORDER BY s.captured_at) AS prev_status
    FROM availability_snapshots s
    WHERE s.point_id IN (SELECT point_id FROM disabled_now)
),
transition_starts AS (
    SELECT DISTINCT ON (point_id) point_id, captured_at AS started_at
    FROM runs
    WHERE prev_status = 'Operating' AND status = 'Disabled'
    ORDER BY point_id, captured_at DESC
),
first_seen AS (
    SELECT point_id, MIN(captured_at) AS first_at
    FROM availability_snapshots
    WHERE point_id IN (SELECT point_id FROM disabled_now)
    GROUP BY point_id
)
SELECT
    p.id, p.inpost_id, p.city, p.street, p.building_no,
    p.latitude, p.longitude, p.image_url, p.location_247, p.is_next,
    'Disabled'::TEXT                                       AS current_status,
    COALESCE(ts.started_at, fs.first_at)                   AS started_at,
    EXTRACT(EPOCH FROM (NOW() - COALESCE(ts.started_at, fs.first_at)))::BIGINT AS duration_seconds
FROM disabled_now dn
JOIN points p          ON p.id = dn.point_id
LEFT JOIN transition_starts ts ON ts.point_id = dn.point_id
LEFT JOIN first_seen   fs ON fs.point_id = dn.point_id
ORDER BY duration_seconds DESC;
`

func (r *AnalyticsRepo) CurrentOutages(ctx context.Context) ([]domain.OngoingOutage, error) {
	rows, err := r.DB.Query(ctx, outagesSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.OngoingOutage
	for rows.Next() {
		var o domain.OngoingOutage
		if err := rows.Scan(
			&o.Locker.ID, &o.Locker.InpostID, &o.Locker.City, &o.Locker.Street, &o.Locker.BuildingNo,
			&o.Locker.Latitude, &o.Locker.Longitude, &o.Locker.ImageURL, &o.Locker.Location247, &o.Locker.IsNext,
			&o.Locker.CurrentStatus,
			&o.StartedAt, &o.DurationSeconds,
		); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

const worstOffendersSQL = `
WITH window_snaps AS (
    SELECT point_id, captured_at, status
    FROM availability_snapshots
    WHERE captured_at >= NOW() - ($1::INT || ' days')::INTERVAL
),
gaps AS (
    SELECT
        point_id,
        status,
        captured_at,
        LAG(status)      OVER (PARTITION BY point_id ORDER BY captured_at) AS prev_status,
        LEAD(captured_at) OVER (PARTITION BY point_id ORDER BY captured_at) AS next_at
    FROM window_snaps
),
downtime AS (
    SELECT
        point_id,
        SUM(EXTRACT(EPOCH FROM (COALESCE(next_at, NOW()) - captured_at)))::BIGINT AS downtime_seconds,
        COUNT(*) FILTER (WHERE prev_status = 'Operating' AND status = 'Disabled') AS events,
        COUNT(*)                                                                  AS samples,
        COUNT(*) FILTER (WHERE status = 'Operating')                              AS operating
    FROM gaps
    WHERE status = 'Disabled'
    GROUP BY point_id
),
per_point AS (
    SELECT
        ws.point_id,
        COALESCE(d.downtime_seconds, 0)                              AS downtime_seconds,
        COALESCE(d.events, 0)                                        AS events,
        AVG((ws.status = 'Operating')::INT::FLOAT)                   AS uptime
    FROM window_snaps ws
    LEFT JOIN downtime d ON d.point_id = ws.point_id
    GROUP BY ws.point_id, d.downtime_seconds, d.events
)
SELECT
    p.id, p.inpost_id, p.city, p.street, p.building_no,
    p.latitude, p.longitude, p.image_url, p.location_247, p.is_next,
    'Disabled'::TEXT,                  -- current_status not relevant in ranking
    NULL::TIMESTAMPTZ,                 -- last_change_at
    pp.events,
    pp.downtime_seconds,
    pp.uptime
FROM per_point pp
JOIN points p ON p.id = pp.point_id
WHERE pp.downtime_seconds > 0
ORDER BY pp.downtime_seconds DESC
LIMIT $2;
`

func (r *AnalyticsRepo) WorstOffenders(ctx context.Context, days, limit int) ([]domain.WorstOffender, error) {
	if days <= 0 {
		days = 7
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	rows, err := r.DB.Query(ctx, worstOffendersSQL, days, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.WorstOffender
	for rows.Next() {
		var w domain.WorstOffender
		if err := rows.Scan(
			&w.Locker.ID, &w.Locker.InpostID, &w.Locker.City, &w.Locker.Street, &w.Locker.BuildingNo,
			&w.Locker.Latitude, &w.Locker.Longitude, &w.Locker.ImageURL, &w.Locker.Location247, &w.Locker.IsNext,
			&w.Locker.CurrentStatus, &w.Locker.LastChangeAt,
			&w.OutageEvents, &w.DowntimeSeconds, &w.UptimePct,
		); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

const outagesTimelineSQL = `
WITH days AS (
    SELECT generate_series(
        date_trunc('day', NOW() - ($1::INT - 1 || ' days')::INTERVAL),
        date_trunc('day', NOW()),
        '1 day'::INTERVAL
    )::DATE AS day
),
runs AS (
    SELECT
        captured_at,
        status,
        LAG(status) OVER (PARTITION BY point_id ORDER BY captured_at) AS prev_status
    FROM availability_snapshots
    WHERE captured_at >= NOW() - ($1::INT || ' days')::INTERVAL
),
events AS (
    SELECT date_trunc('day', captured_at)::DATE AS day, COUNT(*) AS events
    FROM runs
    WHERE prev_status = 'Operating' AND status = 'Disabled'
    GROUP BY day
)
SELECT d.day, COALESCE(e.events, 0)::INT
FROM days d
LEFT JOIN events e ON e.day = d.day
ORDER BY d.day;
`

func (r *AnalyticsRepo) OutagesTimeline(ctx context.Context, days int) ([]domain.DailyOutages, error) {
	if days <= 0 || days > 60 {
		days = 14
	}
	rows, err := r.DB.Query(ctx, outagesTimelineSQL, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.DailyOutages
	for rows.Next() {
		var d domain.DailyOutages
		if err := rows.Scan(&d.Day, &d.OutageEvents); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

const uptimeDistributionSQL = `
WITH per_locker AS (
    SELECT
        point_id,
        AVG((status = 'Operating')::INT::FLOAT) AS uptime
    FROM availability_snapshots
    WHERE captured_at >= NOW() - ($1::INT || ' days')::INTERVAL
    GROUP BY point_id
),
bucketed AS (
    SELECT CASE
        WHEN uptime < 0.50 THEN '<50%'
        WHEN uptime < 0.75 THEN '50-75%'
        WHEN uptime < 0.90 THEN '75-90%'
        WHEN uptime < 0.95 THEN '90-95%'
        WHEN uptime < 0.99 THEN '95-99%'
        ELSE                       '99-100%'
    END AS label
    FROM per_locker
)
SELECT label, COUNT(*)::INT
FROM bucketed
GROUP BY label
ORDER BY CASE label
    WHEN '<50%'    THEN 1
    WHEN '50-75%'  THEN 2
    WHEN '75-90%'  THEN 3
    WHEN '90-95%'  THEN 4
    WHEN '95-99%'  THEN 5
    WHEN '99-100%' THEN 6
END;
`

func (r *AnalyticsRepo) UptimeDistribution(ctx context.Context, days int) ([]domain.UptimeBucket, error) {
	if days <= 0 {
		days = 7
	}
	rows, err := r.DB.Query(ctx, uptimeDistributionSQL, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.UptimeBucket
	for rows.Next() {
		var b domain.UptimeBucket
		if err := rows.Scan(&b.Label, &b.Lockers); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

const networkStatsSQL = `
WITH window_snaps AS (
    SELECT point_id, captured_at, status
    FROM availability_snapshots
    WHERE captured_at >= NOW() - ($1::INT || ' days')::INTERVAL
),
gaps AS (
    SELECT
        point_id, status, captured_at,
        LAG(status)       OVER (PARTITION BY point_id ORDER BY captured_at) AS prev_status,
        LEAD(captured_at) OVER (PARTITION BY point_id ORDER BY captured_at) AS next_at
    FROM window_snaps
)
SELECT
    $1::INT                                                                   AS window_days,
    COALESCE((SELECT COUNT(*)::INT
              FROM gaps
              WHERE prev_status = 'Operating' AND status = 'Disabled'), 0)    AS outage_events,
    COALESCE((SELECT AVG(EXTRACT(EPOCH FROM (COALESCE(next_at, NOW()) - captured_at)))::BIGINT
              FROM gaps
              WHERE status = 'Disabled'), 0)                                  AS avg_outage_duration_seconds,
    COALESCE((SELECT AVG((status = 'Operating')::INT::FLOAT)
              FROM window_snaps), 1)                                          AS network_availability_pct;
`

const lockerUptime7dSQL = `
WITH window_snaps AS (
    SELECT point_id, captured_at, status
    FROM availability_snapshots
    WHERE captured_at >= NOW() - INTERVAL '7 days'
),
runs AS (
    SELECT
        point_id, status,
        LAG(status) OVER (PARTITION BY point_id ORDER BY captured_at) AS prev_status
    FROM window_snaps
)
SELECT
    p.id,
    COALESCE(AVG((ws.status = 'Operating')::INT::FLOAT), 1) AS uptime,
    COALESCE(
        (SELECT COUNT(*) FROM runs
         WHERE runs.point_id = p.id
           AND prev_status = 'Operating' AND status = 'Disabled'), 0
    )::INT AS events
FROM points p
LEFT JOIN window_snaps ws ON ws.point_id = p.id
GROUP BY p.id;
`

func (r *AnalyticsRepo) LockerUptime7d(ctx context.Context) (map[int64]domain.LockerUptime, error) {
	rows, err := r.DB.Query(ctx, lockerUptime7dSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int64]domain.LockerUptime)
	for rows.Next() {
		var id int64
		var u domain.LockerUptime
		if err := rows.Scan(&id, &u.UptimePct, &u.OutageEvents); err != nil {
			return nil, err
		}
		out[id] = u
	}
	return out, rows.Err()
}

func (r *AnalyticsRepo) NetworkStats(ctx context.Context, days int) (domain.NetworkStats, error) {
	if days <= 0 {
		days = 7
	}
	var s domain.NetworkStats
	err := r.DB.QueryRow(ctx, networkStatsSQL, days).Scan(
		&s.WindowDays, &s.OutageEvents, &s.AvgOutageDurationSeconds, &s.NetworkAvailabilityPct,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		s.WindowDays = days
		return s, nil
	}
	return s, err
}
