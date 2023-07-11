create extension if not exists "postgis";
create extension if not exists "timescaledb";

CREATE TABLE devices (
	id BIGINT NOT NULL GENERATED ALWAYS AS IDENTITY,
	code varchar NOT NULL,
	description varchar NOT NULL DEFAULT '',
	organisation varchar NOT NULL,
	properties json NOT NULL DEFAULT '{}'::json,
	"location" geography NULL,
	location_description varchar NOT NULL DEFAULT '',
	altitude DOUBLE PRECISION NULL,
	state SMALLINT NOT NULL DEFAULT 0,
	created_at timestamptz(0) NOT NULL DEFAULT(NOW() AT TIME ZONE 'UTC'),

    PRIMARY KEY(id)
);


CREATE TABLE sensors (
	id BIGINT NOT NULL GENERATED ALWAYS AS IDENTITY,
	code varchar NOT NULL,
	device_id BIGINT NOT NULL REFERENCES devices(id),
	description varchar NOT NULL DEFAULT '',
	external_id varchar NULL,
	properties json NOT NULL DEFAULT '{}'::json,
	archive_time INT NULL,
	brand varchar NOT NULL DEFAULT '',
	created_at timestamptz(0) NOT NULL DEFAULT(NOW() AT TIME ZONE 'UTC'),

	PRIMARY KEY (id)
);

CREATE TABLE datastreams (
	id uuid NOT NULL,
	description varchar NOT NULL DEFAULT '',
	sensor_id BIGINT NOT NULL,
	observed_property varchar NOT NULL,
	unit_of_measurement varchar NOT NULL,
	created_at timestamptz(0) NOT NULL DEFAULT(NOW() AT TIME ZONE 'UTC'),

	PRIMARY KEY (id),
	UNIQUE (sensor_id, observed_property)
);

CREATE TABLE measurements (
	id BIGINT NOT NULL GENERATED ALWAYS AS IDENTITY,
	uplink_message_id uuid NULL,
	device_id BIGINT NOT NULL,
	device_code varchar NOT NULL,
	device_description varchar NOT NULL DEFAULT '',
	measurement_timestamp timestamptz(0) NOT NULL,
	measurement_value float8 NOT NULL,
	coordinates geography(point, 4326) NULL,
	location_id BIGINT NULL,
	device_location_description varchar NULL,
	measurement_location geography(point, 4326) NULL,
	sensor_code varchar NULL,
	sensor_description varchar NULL,
	sensor_external_id varchar NULL,
	measurement_properties jsonb NOT NULL DEFAULT '{}'::jsonb,
	organisation_id BIGINT NULL,
	organisation_name varchar NULL,
	organisation_address varchar NULL,
	organisation_zipcode varchar NULL,
	organisation_city varchar NULL,
	organisation_chamber_of_commerce_id varchar NULL,
	organisation_headquarter_id varchar NULL,
	device_location geography(point, 4326) NULL,
	device_properties jsonb NULL,
	sensor_id BIGINT NULL,
	sensor_properties jsonb NULL,
	sensor_brand varchar NULL,
	organisation_state SMALLINT NOT NULL,
	organisation_archive_time SMALLINT NULL,
	device_altitude DOUBLE PRECISION NULL,
	device_state SMALLINT NOT NULL,
	sensor_archive_time SMALLINT NULL,
	datastream_id varchar NOT NULL,
	datastream_description varchar NULL,
	datastream_observed_property varchar NULL,
	datastream_unit_of_measurement varchar NULL,
	measurement_altitude DOUBLE PRECISION NULL,
	measurement_expiration DATE NOT NULL DEFAULT (now() + '7 days'::interval),
	created_at timestamptz(0) NOT NULL DEFAULT(NOW() AT TIME ZONE 'UTC'),

	PRIMARY KEY (id, measurement_timestamp)
);
SELECT create_hypertable('measurements', 'measurement_timestamp');


CREATE TABLE pipelines (
	id uuid NOT NULL,
	description varchar NULL DEFAULT '',
	status varchar NOT NULL,
	last_status_change timestamptz NOT NULL,
	created_at timestamptz(0) NOT NULL DEFAULT(NOW() AT TIME ZONE 'UTC'),

	PRIMARY KEY (id)
);


CREATE TABLE pipeline_steps (
	pipeline_id uuid NOT NULL REFERENCES pipelines(id),
	pipeline_step INT NOT NULL,
	image varchar NOT NULL,

	UNIQUE (pipeline_id, pipeline_step)
);

