-- Enable the uuid-ossp extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Insert random data into the table with archived_at within the last 24 hours
INSERT INTO archived_ingress_dtos (tracing_id, raw_message, dto_owner_id, dto_pipeline_id, dto_payload, dto_created_at, archived_at, expires_at)
SELECT
    uuid_generate_v4() AS tracing_id,
    E'\\x' || substring(md5(random()::text), 1, 32)::bytea AS raw_message,
    floor(random() * 1000)::bigint AS dto_owner_id,
    md5(random()::text)::uuid::text AS dto_pipeline_id,
    E'\\x' || substring(md5(random()::text), 1, 32)::bytea AS dto_payload,
    NOW() - (random() * interval '365 days') AS dto_created_at,
    NOW() - (interval '1 day' * random()) AS archived_at,
    NOW() + (random() * interval '365 days') AS expires_at
FROM generate_series(1, 1000);