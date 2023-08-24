BEGIN;

    CREATE TABLE archived_ingress_dtos (
        tracing_id VARCHAR NOT NULL,
        raw_message BYTEA,


        dto_owner_id BIGINT,
        dto_pipeline_id VARCHAR,
        dto_payload BYTEA,
        dto_created_at TIMESTAMPTZ,

        archived_at TIMESTAMPTZ DEFAULT(now()),
        expires_at TIMESTAMPTZ NOT NULL
    );

COMMIT;
