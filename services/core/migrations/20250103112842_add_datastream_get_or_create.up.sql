CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

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
        uuid_generate_v4(), arg_tenant_id, arg_sensor_id, arg_observed_property, arg_unit_of_measurement
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
