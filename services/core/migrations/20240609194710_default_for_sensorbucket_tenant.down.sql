BEGIN;

    DROP INDEX IF EXISTS pipelines_idx_tenant_id;
    DROP INDEX IF EXISTS datastreams_idx_tenant_id;
    DROP INDEX IF EXISTS sensor_groups_idx_tenant_id;

COMMIT;
