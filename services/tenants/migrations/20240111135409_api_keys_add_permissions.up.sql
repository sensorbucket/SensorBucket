BEGIN;

CREATE TABLE api_key_permissions (
    id SERIAL PRIMARY KEY,
    permission VARCHAR(255) NOT NULL,
    api_key_id BIGINT REFERENCES api_keys(id) ON DELETE CASCADE
);

-- Add an index on api_key_id for performance optimization
CREATE INDEX idx_api_key_permissions_api_key_id ON api_key_permissions(api_key_id);

COMMIT;