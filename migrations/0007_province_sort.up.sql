CREATE INDEX IF NOT EXISTS idx_points_province ON points(province);

ALTER TABLE coverage_grids RENAME COLUMN city TO province;
ALTER TABLE coverage_recommendations RENAME COLUMN city TO province;
