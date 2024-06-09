BEGIN;

    ALTER TABLE sensor_groups DROP COLUMN IF EXISTS tenant_id;
    ALTER TABLE datastreams DROP COLUMN IF EXISTS tenant_id;
    ALTER TABLE pipelines DROP COLUMN IF EXISTS tenant_id;
    
    DROP INDEX IF EXISTS sensors_idx_tenant_id;
    ALTER TABLE sensors DROP COLUMN IF EXISTS tenant_id;

    DROP INDEX IF EXISTS devices_idx_tenant_id;
    ALTER TABLE devices DROP COLUMN IF EXISTS tenant_id;
    ALTER TABLE devices ADD COLUMN organisation varchar;

COMMIT;
