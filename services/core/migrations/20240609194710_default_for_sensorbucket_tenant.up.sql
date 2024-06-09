BEGIN;

    UPDATE devices SET tenant_id = 1 WHERE tenant_id = 0;
    UPDATE pipelines SET tenant_id = 1 WHERE tenant_id = 0;
    UPDATE datastreams SET tenant_id = 1 WHERE tenant_id = 0;

    ALTER TABLE sensor_groups ADD COLUMN tenant_id BIGINT NOT NULL DEFAULT 1;
    ALTER TABLE sensor_groups ALTER COLUMN tenant_id DROP DEFAULT;

    ALTER TABLE pipelines ALTER COLUMN tenant_id DROP DEFAULT;
    CREATE INDEX pipelines_idx_tenant_id ON pipelines (tenant_id);

    ALTER TABLE datastreams ALTER COLUMN tenant_id DROP DEFAULT;
    CREATE INDEX datastreams_idx_tenant_id ON datastreams (tenant_id);

    ALTER TABLE sensor_groups ALTER COLUMN tenant_id DROP DEFAULT;
    CREATE INDEX sensor_groups_idx_tenant_id ON sensor_groups (tenant_id);

COMMIT;
