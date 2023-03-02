BEGIN;

ALTER TABLE "measurements" RENAME "organisation_coc" TO "organisation_chamber_of_commerce_id";
ALTER TABLE "measurements" RENAME "organisation_location_coc" TO "organisation_headquarter_id";

ALTER TABLE "measurements" ADD "organisation_state" INT NOT NULL;
ALTER TABLE "measurements" ADD "organisation_archive_time" INT;
ALTER TABLE "measurements" ADD "device_altitude" REAL;
ALTER TABLE "measurements" ADD "device_state" INT NOT NULL;
ALTER TABLE "measurements" ADD "sensor_archive_time" INT;
ALTER TABLE "measurements" ADD "datastream_id" VARCHAR NOT NULL;
ALTER TABLE "measurements" ADD "datastream_description" VARCHAR;
ALTER TABLE "measurements" ADD "datastream_observed_property" VARCHAR:
ALTER TABLE "measurements" ADD "datastream_unit_of_measurement" VARCHAR;
ALTER TABLE "measurements" ADD "measurement_altitude" REAL;

ALTER TABLE "measurement" DROP "measurement_type";
ALTER TABLE "measurement" DROP "measurement_unit";
ALTER TABLE "measurement" DROP "measurement_value_prefix";
ALTER TABLE "measurement" DROP "measurement_value_factor";

COMMIT;
