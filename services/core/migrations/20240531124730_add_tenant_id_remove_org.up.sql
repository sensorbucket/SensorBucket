BEGIN;
    -- Drop the organisation column
    ALTER TABLE devices
    DROP COLUMN IF EXISTS organisation;

    ALTER TABLE devices ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
    ALTER TABLE devices ALTER COLUMN tenant_id DROP DEFAULT;
    CREATE INDEX devices_idx_tenant_id ON devices (tenant_id);

    ALTER TABLE sensors ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
    ALTER TABLE sensors ALTER COLUMN tenant_id DROP DEFAULT;
    CREATE INDEX sensors_idx_tenant_id ON sensors (tenant_id);

    ALTER TABLE pipelines ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
    ALTER TABLE pipelines ALTER COLUMN tenant_id DROP DEFAULT;
    CREATE INDEX pipelines_idx_tenant_id ON pipelines (tenant_id);

    ALTER TABLE datastreams ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
    ALTER TABLE datastreams ALTER COLUMN tenant_id DROP DEFAULT;
    CREATE INDEX datastreams_idx_tenant_id ON datastreams (tenant_id);

    ALTER TABLE sensor_groups ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
    ALTER TABLE sensor_groups ALTER COLUMN tenant_id DROP DEFAULT;
    CREATE INDEX sensor_groups_idx_tenant_id ON sensor_groups (tenant_id);

COMMIT;

