ALTER TABLE points
    ALTER COLUMN inpost_id     TYPE VARCHAR(50),
    ALTER COLUMN country       TYPE VARCHAR(10),
    ALTER COLUMN status        TYPE VARCHAR(40),
    ALTER COLUMN location_type TYPE VARCHAR(50),
    ALTER COLUMN physical_type TYPE VARCHAR(50),
    ALTER COLUMN building_no   TYPE VARCHAR(60);

ALTER TABLE availability_snapshots
    ALTER COLUMN status TYPE VARCHAR(40);
