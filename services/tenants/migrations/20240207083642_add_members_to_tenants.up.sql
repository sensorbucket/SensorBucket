BEGIN;

CREATE TABLE tenant_members (
    id bigint primary key not null,
    tenant_id bigint not null,
    user_id varchar not null,

    FOREIGN KEY(tenant_id) REFERENCES tenants(id)
);

CREATE TABLE tenant_member_permissions (
    member_id bigint not null,
    permission varchar not null,

    FOREIGN KEY(member_id) REFERENCES tenant_members(id),
    PRIMARY KEY(member_id, permission)
);

COMMIT;
