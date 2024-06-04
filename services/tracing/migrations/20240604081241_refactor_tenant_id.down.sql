BEGIN;
ALTER TABLE archived_ingress_dtos RENAME COLUMN dto_tenant_id TO dto_owner_id;
COMMIT;
