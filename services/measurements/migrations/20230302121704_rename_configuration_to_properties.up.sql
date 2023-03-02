BEGIN;

ALTER TABLE "measurements" RENAME COLUMN "measurement_metadata" TO "measurement_properties";
ALTER TABLE "measurements" RENAME COLUMN "sensor_config" TO "sensor_properties";
ALTER TABLE "measurements" RENAME COLUMN "device_configuration" TO "device_properties";

COMMIT;
