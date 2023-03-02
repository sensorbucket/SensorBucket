BEGIN;

ALTER TABLE measurements
    ADD COLUMN organisation_id BIGINT,
    ADD COLUMN organisation_name VARCHAR,
    ADD COLUMN organisation_address VARCHAR,
    ADD COLUMN organisation_zipcode VARCHAR,
    ADD COLUMN organisation_city VARCHAR,
    ADD COLUMN organisation_coc VARCHAR,
    ADD COLUMN organisation_location_coc VARCHAR,
    ADD COLUMN device_location GEOGRAPHY(POINT,4326),
    ALTER COLUMN location_name RENAME TO device_location_description,
    ADD COLUMN device_configuration JSONB,
    ADD COLUMN sensor_id BIGINT,
    ADD COLUMN sensor_config JSONB,
    ADD COLUMN sensor_brand VARCHAR,
    ALTER COLUMN measurement_type_unit RENAME TO measurement_unit,
    ALTER COLUMN timestamp RENAME TO measurement_timestamp,
    ALTER COLUMN value RENAME TO measurement_value,
    ADD COLUMN measurement_value_prefix VARCHAR,
    ADD COLUMN measurement_value_factor INT,
    ALTER COLUMN location_coordinates RENAME TO measurement_location,
    ALTER COLUMN metadata RENAME TO measurement_metadata;

COMMIT;
