BEGIN;

CREATE TABLE tenant_members (
    id bigint primary key not null,
    tenant_id bigint not null,
    user_id varchar not null,
    permissions varchar[] not null default([]),

    FOREIGN KEY(tenant_id) REFERENCES tenants(id)
);

COMMIT;
