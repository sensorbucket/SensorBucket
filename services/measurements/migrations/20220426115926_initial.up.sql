CREATE EXTENSION IF NOT EXISTS "postgis";

CREATE TABLE measurements (
  id BIGINT GENERATED ALWAYS AS IDENTITY,
  uplink_message_id UUID,
  device_id BIGINT NOT NULL,
  device_code VARCHAR NOT NULL,
  device_description VARCHAR NOT NULL DEFAULT(''),
  timestamp TIMESTAMPTZ(0) NOT NULL,
  value DOUBLE PRECISION NOT NULL,
  measurement_type VARCHAR NOT NULL,
  measurement_type_unit VARCHAR NOT NULL,
  coordinates GEOGRAPHY(POINT,4326),
  location_id INTEGER,
  location_name VARCHAR,
  location_coordinates GEOGRAPHY(POINT,4326),
  sensor_code VARCHAR,
  sensor_description VARCHAR,
  sensor_external_id VARCHAR,
  metadata JSONB NOT NULL DEFAULT '{}',

  CONSTRAINT measurements_pkey PRIMARY KEY (id, "timestamp")
);
SELECT create_hypertable('measurements', 'timestamp');
