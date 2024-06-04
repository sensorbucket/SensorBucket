BEGIN;
ALTER TABLE archived_ingress_dtos RENAME COLUMN dto_owner_id TO dto_tenant_id;
COMMIT;
