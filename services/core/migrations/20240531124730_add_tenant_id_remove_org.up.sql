BEGIN;
    -- Drop the organisation column
    ALTER TABLE devices
    DROP COLUMN IF EXISTS organisation;

    -- Add the tenant_id column
    ALTER TABLE devices
    ADD COLUMN tenant_id BIGINT;

    -- Set all NULL tenant_id values to 0
    UPDATE devices
    SET tenant_id = 0
    WHERE tenant_id IS NULL;

    -- Make tenant_id NOT NULL
    ALTER TABLE devices
    ALTER COLUMN tenant_id SET NOT NULL;

    -- Create an index on tenant_id
    CREATE INDEX devices_idx_tenant_id ON devices (tenant_id);


    -- Add the tenant_id column
    ALTER TABLE sensors
    ADD COLUMN tenant_id BIGINT;

    -- Set tenant_id to device tenant_id
    UPDATE sensors 
    SET tenant_id = devices.tenant_id
    FROM devices
    WHERE devices.id = sensors.device_id;

    -- Make tenant_id NOT NULL
    ALTER TABLE sensors
    ALTER COLUMN tenant_id SET NOT NULL;

    -- Create an index on tenant_id
    CREATE INDEX sensors_idx_tenant_id ON sensors (tenant_id);


    -- Add tenantID to pipelines
    ALTER TABLE pipelines ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 0;

    -- Add tenantID to datastreams
    ALTER TABLE datastreams ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 0;

COMMIT;

