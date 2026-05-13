CREATE TABLE IF NOT EXISTS points (
    id              SERIAL PRIMARY KEY,
    inpost_id       VARCHAR(20) UNIQUE NOT NULL,
    country         VARCHAR(5) NOT NULL,
    status          VARCHAR(20) NOT NULL,
    latitude        DOUBLE PRECISION NOT NULL,
    longitude       DOUBLE PRECISION NOT NULL,
    city            VARCHAR(100),
    province        VARCHAR(100),
    post_code       VARCHAR(10),
    street          VARCHAR(200),
    building_no     VARCHAR(20),
    location_type   VARCHAR(20),
    is_next         BOOLEAN,
    location_247    BOOLEAN,
    physical_type   VARCHAR(20),
    image_url       TEXT,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_points_city ON points(city);

CREATE TABLE IF NOT EXISTS availability_snapshots (
    id          BIGSERIAL PRIMARY KEY,
    point_id    INTEGER NOT NULL REFERENCES points(id) ON DELETE CASCADE,
    captured_at TIMESTAMPTZ NOT NULL,
    status      VARCHAR(20) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_snap_point_time ON availability_snapshots(point_id, captured_at DESC);
