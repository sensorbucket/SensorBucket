CREATE TABLE measurements (
  thing_urn VARCHAR NOT NULL,
  timestamp TIMESTAMPTZ NOT NULL,
  value DOUBLE PRECISION NOT NULL,
  measurement_type VARCHAR NOT NULL,
  measurement_type_unit VARCHAR NOT NULL,
  location_id INTEGER,
  coordinates GEOGRAPHY(POINT,4326) NOT NULL,
  metadata JSONB,
  
  PRIMARY KEY (thing_urn, timestamp)
);
