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

CREATE TABLE "measurement_types" (
    "name" VARCHAR NOT NULL,

    PRIMARY KEY("name")
);
CREATE TABLE "measurement_units" (
    "name" VARCHAR NOT NULL,

    PRIMARY KEY("name")
);

CREATE TABLE "sensor_measurements" (
    "sensor_type_id" BIGINT REFERENCES "sensor_types"("id"),
    "measurement_type" VARCHAR REFERENCES "measurement_types"("name"),
    "measurement_unit" VARCHAR REFERENCES "measurement_units"("name"),

    UNIQUE("sensor_type_id", "measurement_type", "measurement_unit")
);

ALTER TABLE "sensors"
    ALTER COLUMN "type_id" TYPE BIGINT,
    ALTER COLUMN "goal_id" TYPE BIGINT,
    ALTER COLUMN "type_id" SET NOT NULL,
    ALTER COLUMN "goal_id" SET NOT NULL,
    ADD CONSTRAINT fk_sensor_type FOREIGN KEY ("type_id") REFERENCES "sensor_types"("id"),
    ADD CONSTRAINT fk_sensor_goal FOREIGN KEY ("goal_id") REFERENCES "sensor_goals"("id");

COMMIT;
