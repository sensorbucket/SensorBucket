BEGIN;

CREATE TABLE tenant_members (
    tenant_id bigint not null,
    user_id varchar not null,
    permissions varchar[] not null default('{}'),

    FOREIGN KEY(tenant_id) REFERENCES tenants(id),
    PRIMARY KEY(tenant_id, user_id)
);

COMMIT;
