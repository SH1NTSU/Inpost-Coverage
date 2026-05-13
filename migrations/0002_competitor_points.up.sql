CREATE TABLE IF NOT EXISTS competitor_points (
    id         SERIAL PRIMARY KEY,
    network    VARCHAR(50) NOT NULL,
    name       VARCHAR(200),
    latitude   DOUBLE PRECISION NOT NULL,
    longitude  DOUBLE PRECISION NOT NULL,
    address    VARCHAR(300),
    osm_id     BIGINT,
    raw_tags   JSONB,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(osm_id)
);

CREATE INDEX IF NOT EXISTS idx_comp_network ON competitor_points(network);
CREATE INDEX IF NOT EXISTS idx_comp_geo     ON competitor_points(latitude, longitude);
