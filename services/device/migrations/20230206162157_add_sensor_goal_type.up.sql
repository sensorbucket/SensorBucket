BEGIN;

CREATE TABLE "sensor_types" (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL,
    "description" VARCHAR NOT NULL,

    PRIMARY KEY("id")
);

CREATE TABLE "sensor_goals" (
    "id" BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL,
    "name" VARCHAR NOT NULL,
    "description" VARCHAR DEFAULT(''),

    PRIMARY KEY("id")
);

CREATE TABLE "sensor_measurements" (
    "sensor_type_id" BIGINT REFERENCES "sensor_types"("id"),
    "measurement_type" VARCHAR REFERENCES "measurement_types"("name"),
    "measurement_unit" VARCHAR REFERENCES "measurement_units"("name"),

    UNIQUE("sensor_type_id", "measurement_type", "measurement_unit")
);

ALTER TABLE "sensors"
    ALTER COLUMN "type_id" BIGINT REFERENCES "sensor_types"("id"),
    ALTER COLUMN "goal_id" BIGINT REFERENCES "sensor_goals"("id");

COMMIT;
