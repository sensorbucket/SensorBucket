-- Save extisting measurements, the sysadmin must recover these by performing an export and import:
-- psql 'postgres://<user>:<password>@<host>:<port>/<database>' -c "/COPY (select * from measurements) TO '<path>/measurements.csv' WITH csv"
-- Import using psql if less than a few million rows, otherwise use timescaledb-parallel-copy tool
-- psql 'postgres://<user>:<password>@<host>:<port>/<database>' -c "/COPY measurements FROM '<path>/measurements.csv'"
-- Make sure to drop the old table if recovery was succesful
alter table measurements rename to measurements_backup;
alter index measurements_measurement_timestamp_idx rename to measurements_backup_measurement_timestamp_idx;
alter index measurements_pkey rename to measurements_backup_pkey;

-- Unused as the table already has: device_location and measurement_location
alter table measurements drop column coordinates;

-- Recreate the table with new index and
CREATE TABLE measurements (
	id BIGINT NOT NULL GENERATED ALWAYS AS IDENTITY,
	uplink_message_id UUID NULL,
	device_id BIGINT NOT NULL,
	device_code TEXT NOT NULL,
	device_description TEXT NOT NULL DEFAULT '',
	measurement_timestamp TIMESTAMPTZ(0) NOT NULL,
	measurement_value FLOAT8 NOT NULL,
	coordinates GEOGRAPHY(point, 4326) NULL,
	location_id BIGINT NULL,
	device_location_description TEXT NULL,
	measurement_location GEOGRAPHY(point, 4326) NULL,
	sensor_code TEXT NULL,
	sensor_description TEXT NULL,
	sensor_external_id TEXT NULL,
	measurement_properties JSONB NOT NULL DEFAULT '{}'::jsonb,
	organisation_id BIGINT NULL,
	organisation_name TEXT NULL,
	organisation_address TEXT NULL,
	organisation_zipcode TEXT NULL,
	organisation_city TEXT NULL,
	organisation_chamber_of_commerce_id TEXT NULL,
	organisation_headquarter_id TEXT NULL,
	device_location GEOGRAPHY(point, 4326) NULL,
	device_properties JSONB NULL,
	sensor_id BIGINT NULL,
	sensor_properties JSONB NULL,
	sensor_brand TEXT NULL,
	organisation_state SMALLINT NOT NULL,
	organisation_archive_time SMALLINT NULL,
	device_altitude DOUBLE PRECISION NULL,
	device_state SMALLINT NOT NULL,
	sensor_archive_time SMALLINT NULL,
	datastream_id UUID NOT NULL,
	datastream_description TEXT NULL,
	datastream_observed_property TEXT NULL,
	datastream_unit_of_measurement TEXT NULL,
	measurement_altitude DOUBLE PRECISION NULL,
	measurement_expiration DATE NOT NULL DEFAULT (NOW() + '7 days'::interval),
	created_at TIMESTAMPTZ(0) NOT NULL DEFAULT(NOW() AT TIME ZONE 'UTC'),

	PRIMARY KEY (datastream_id, measurement_timestamp, id)
);
SELECT create_hypertable('measurements', 'measurement_timestamp');
SELECT add_dimension('measurements', 'datastream_id', number_partitions => 50);
CREATE INDEX measurements_query_idx ON measurements(datastream_id, measurement_timestamp DESC);

--
-- Update the datastream get or create function to use a polyfilled UUIDv7 function
--

-- Thanks: https://postgresql.verite.pro/blog/2024/07/15/uuid-v7-pure-sql.html
CREATE OR REPLACE FUNCTION uuid_generate_v7() RETURNS uuid
AS $$
  -- Replace the first 48 bits of a uuidv4 with the current
  -- number of milliseconds since 1970-01-01 UTC
  -- and set the "ver" field to 7 by setting additional bits
  select encode(
    set_bit(
      set_bit(
        overlay(uuid_send(uuid_generate_v4()) placing
	  substring(int8send((extract(epoch from clock_timestamp())*1000)::bigint) from 3)
	  from 1 for 6),
	52, 1),
      53, 1), 'hex')::uuid;
$$ LANGUAGE sql volatile;

-- Update to use uuid v7
CREATE OR REPLACE FUNCTION find_or_create_datastream(
  arg_tenant_id datastreams.tenant_id%TYPE,
  arg_sensor_id datastreams.sensor_id%TYPE,
  arg_observed_property datastreams.observed_property%TYPE,
  arg_unit_of_measurement datastreams.unit_of_measurement%TYPE
)
RETURNS SETOF datastreams AS $$
DECLARE
  return_datastreams datastreams%ROWTYPE;
BEGIN
  SELECT 
    id, description, sensor_id, observed_property, unit_of_measurement, created_at, tenant_id
  INTO return_datastreams FROM datastreams WHERE 
    tenant_id = arg_tenant_id
    AND observed_property = arg_observed_property
    AND sensor_id = arg_sensor_id
    AND unit_of_measurement = arg_unit_of_measurement;
  IF FOUND THEN
    RETURN NEXT return_datastreams;
  ELSE
    BEGIN
      RETURN QUERY INSERT INTO datastreams (
        id, tenant_id, sensor_id, observed_property, unit_of_measurement
      ) VALUES (
        uuid_generate_v7(), arg_tenant_id, arg_sensor_id, arg_observed_property, arg_unit_of_measurement
      ) RETURNING 
          id, description, sensor_id, observed_property, 
          unit_of_measurement, created_at, tenant_id;
    EXCEPTION WHEN unique_violation THEN
      RETURN QUERY SELECT 
          id, description, sensor_id, observed_property, unit_of_measurement,
          created_at, tenant_id
        FROM datastreams WHERE 
          tenant_id = tenant_id
          AND observed_property = observed_property
          AND sensor_id = sensor_id
          AND unit_of_measurement = unit_of_measurement;
    END;
  END IF;
END;
$$ LANGUAGE plpgsql;
