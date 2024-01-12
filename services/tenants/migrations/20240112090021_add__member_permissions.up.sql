BEGIN;

CREATE TABLE IF NOT EXISTS tenant_permissions (
    id SERIAL PRIMARY KEY,
    permission VARCHAR(255) NOT NULL,
    tenant_id BIGINT REFERENCES tenants(id) NOT NULL
);

CREATE TABLE IF NOT EXISTS member_permissions (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    created TIMESTAMP NOT NULL,

    -- Ensure all member permissions are also deleted when a tenant permission is deleted
    permission_id BIGINT REFERENCES tenant_permissions(id) ON DELETE CASCADE NOT NULL
);

-- Add an index performance optimization
CREATE INDEX idx_tenant_permissions_tenant_id ON tenant_permissions(tenant_id);
CREATE INDEX idx_member_permissions_tenant_id ON member_permissions(permission_id);

COMMIT;