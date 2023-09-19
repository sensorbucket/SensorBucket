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

CREATE TABLE steps (
    tracing_id VARCHAR(255) NOT NULL,
    step_index BIGINT NOT NULL,
    steps_remaining BIGINT NOT NULL,
    start_time timestamptz(3) NOT NULL,
    device_id BIGINT,
    error TEXT,
    PRIMARY KEY (tracing_id, step_index)
);

COMMIT;

CREATE
OR REPLACE VIEW enriched_steps_view AS (
    SELECT
        currentstep.*,
        (
            SELECT (EXTRACT(EPOCH FROM (nextstep.start_time - currentstep.start_time)) * 1000)::bigint
        ) AS duration,
        (
            CASE
                WHEN currentstep.error <> '' THEN 5
                WHEN currentstep.steps_remaining <> 0
                AND nextstep IS NULL THEN 4
                ELSE 3
            END
        ) AS status,
        MAX(
            CASE
                WHEN currentstep.error <> '' THEN 5
                WHEN currentstep.steps_remaining <> 0
                AND nextstep IS NULL THEN 4
                ELSE 3
            END
        ) OVER (PARTITION BY currentstep.tracing_id) AS trace_status
    FROM
        (
            SELECT
                steps.*
            FROM
                (
                    SELECT
                        DISTINCT tracing_id
                    FROM
                        steps
                ) trace_ids
                LEFT JOIN steps steps ON trace_ids.tracing_id = steps.tracing_id
        ) currentStep
        LEFT JOIN steps nextstep ON currentstep.tracing_id = nextstep.tracing_id
        AND nextstep.step_index = currentstep.step_index + 1
);
