BEGIN;
ALTER TABLE user_workers RENAME COLUMN organisation TO tenant_id;
COMMIT;

