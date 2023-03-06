BEGIN;

CREATE TABLE "datastreams" (
    "id" UUID NOT NULL PRIMARY KEY,
    "description" VARCHAR NOT NULL DEFAULT(''),
    "sensor_id" BIGINT NOT NULL,
    "observed_property" VARCHAR NOT NULL,
    "unit_of_measurement" VARCHAR NOT NULL,

    UNIQUE ("sensor_id", "observed_property")
);

COMMIT;
