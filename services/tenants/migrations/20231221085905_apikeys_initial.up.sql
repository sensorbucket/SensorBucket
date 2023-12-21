BEGIN;

CREATE TABLE IF NOT EXISTS api_keys (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    value VARCHAR(255) NOT NULL,
    created TIMESTAMP NOT NULL,
    expiration_date TIMESTAMP,
    tenant_id INT NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

COMMIT;