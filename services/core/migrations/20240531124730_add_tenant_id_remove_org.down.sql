BEGIN;

    ALTER TABLE sensor_groups DROP COLUMN IF EXISTS tenant_id;
    ALTER TABLE datastreams DROP COLUMN IF EXISTS tenant_id;
    ALTER TABLE pipelines DROP COLUMN IF EXISTS tenant_id;
    ALTER TABLE sensors DROP COLUMN IF EXISTS tenant_id;
    ALTER TABLE devices DROP COLUMN IF EXISTS tenant_id;
    ALTER TABLE devices ADD COLUMN organisation varchar;

COMMIT;
