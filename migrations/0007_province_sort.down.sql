ALTER TABLE coverage_recommendations RENAME COLUMN province TO city;
ALTER TABLE coverage_grids RENAME COLUMN province TO city;

DROP INDEX IF EXISTS idx_points_province;
