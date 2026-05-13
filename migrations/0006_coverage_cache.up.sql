CREATE TABLE IF NOT EXISTS coverage_grids (
    city        VARCHAR(120) NOT NULL,
    cell_meters INTEGER NOT NULL,
    version     VARCHAR(32) NOT NULL,
    summary     JSONB NOT NULL,
    cells       JSONB NOT NULL,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (city, cell_meters, version)
);

CREATE TABLE IF NOT EXISTS coverage_recommendations (
    city        VARCHAR(120) NOT NULL,
    limit_count INTEGER NOT NULL,
    version     VARCHAR(32) NOT NULL,
    payload     JSONB NOT NULL,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (city, limit_count, version)
);
