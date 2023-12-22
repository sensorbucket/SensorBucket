BEGIN;

CREATE TABLE IF NOT EXISTS tenants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(75) NOT NULL,
    address VARCHAR(50) NOT NULL,
    zip_code VARCHAR(7) NOT NULL,
    city VARCHAR(50) NOT NULL,
    chamber_of_commerce_id VARCHAR NOT NULL,
    headquarter_id VARCHAR NOT NULL,
    archive_time BIGINT,
    state INTEGER NOT NULL,
    logo VARCHAR(255),
    created TIMESTAMP NOT NULL,
    parent_tenant_id  INTEGER REFERENCES tenants(id)
);

COMMIT;