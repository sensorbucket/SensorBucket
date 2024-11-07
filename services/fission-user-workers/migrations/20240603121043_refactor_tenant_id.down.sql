BEGIN;
ALTER TABLE user_workers RENAME COLUMN tenant_id TO organisation;
COMMIT;
