BEGIN;

-- Drop the tenant_id index
DROP INDEX IF EXISTS devices_idx_tenant_id;
DROP INDEX IF EXISTS sensors_idx_tenant_id;

-- Remove the tenant_id column
ALTER TABLE devices
DROP COLUMN IF EXISTS tenant_id;

-- Re-add the organisation column (assuming it was of type TEXT)
ALTER TABLE devices
ADD COLUMN organisation varchar;

-- Set all NULL organisations values to ""
UPDATE devices
SET orgnaisation = ''
WHERE organisation IS NULL;

-- Make organisation NOT NULL
ALTER TABLE devices
ALTER COLUMN organisation SET NOT NULL;

-- Drop tenant_id on sensors
ALTER TABLE sensors DROP COLUMN IF EXISTS tenant_id;

COMMIT;
