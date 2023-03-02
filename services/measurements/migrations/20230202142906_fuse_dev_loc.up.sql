BEGIN;

ALTER TABLE "measurements"
    ADD COLUMN "organisation_id" BIGINT,
    ADD COLUMN "organisation_name" VARCHAR,
    ADD COLUMN "organisation_address" VARCHAR,
    ADD COLUMN "organisation_zipcode" VARCHAR,
    ADD COLUMN "organisation_city" VARCHAR,
    ADD COLUMN "organisation_coc" VARCHAR,
    ADD COLUMN "organisation_location_coc" VARCHAR,
    ADD COLUMN "device_location" GEOGRAPHY(POINT,4326),
    ADD COLUMN "device_configuration" JSONB,
    ADD COLUMN "sensor_id" BIGINT,
    ADD COLUMN "sensor_config" JSONB,
    ADD COLUMN "sensor_brand" VARCHAR,
    ADD COLUMN "measurement_value_prefix" VARCHAR,
    ADD COLUMN "measurement_value_factor" INT;

ALTER TABLE "measurements" RENAME COLUMN "location_name" TO "device_location_description";
ALTER TABLE "measurements" RENAME COLUMN "measurement_type_unit" TO "measurement_unit";
ALTER TABLE "measurements" RENAME COLUMN "timestamp" TO "measurement_timestamp";
ALTER TABLE "measurements" RENAME COLUMN "value" TO "measurement_value";
ALTER TABLE "measurements" RENAME COLUMN "location_coordinates" TO "measurement_location";
ALTER TABLE "measurements" RENAME COLUMN "metadata" TO "measurement_metadata";

COMMIT;
