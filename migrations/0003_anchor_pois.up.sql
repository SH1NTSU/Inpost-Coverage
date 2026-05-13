CREATE TABLE IF NOT EXISTS anchor_pois (
    id         SERIAL PRIMARY KEY,
    poi_type   VARCHAR(40) NOT NULL,
    brand      VARCHAR(120),
    name       VARCHAR(200),
    latitude   DOUBLE PRECISION NOT NULL,
    longitude  DOUBLE PRECISION NOT NULL,
    address    VARCHAR(300),
    osm_id     BIGINT UNIQUE NOT NULL,
    raw_tags   JSONB,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_anchor_type  ON anchor_pois(poi_type);
CREATE INDEX IF NOT EXISTS idx_anchor_geo   ON anchor_pois(latitude, longitude);
CREATE INDEX IF NOT EXISTS idx_anchor_brand ON anchor_pois(brand);
