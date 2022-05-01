CREATE EXTENSION IF NOT EXISTS "postgis";

CREATE TABLE measurements (
  id BIGINT GENERATED ALWAYS AS IDENTITY,
  thing_urn VARCHAR NOT NULL,
  timestamp TIMESTAMPTZ(0) NOT NULL,
  value DOUBLE PRECISION NOT NULL,
  measurement_type VARCHAR NOT NULL,
  measurement_type_unit VARCHAR NOT NULL,
  location_id INTEGER,
  coordinates GEOGRAPHY(POINT,4326) NOT NULL,
  metadata JSONB NOT NULL DEFAULT '{}',

  CONSTRAINT measurements_pkey PRIMARY KEY (id, "timestamp")
);
SELECT create_hypertable('measurements', 'timestamp');