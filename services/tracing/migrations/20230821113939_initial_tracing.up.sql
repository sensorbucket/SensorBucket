BEGIN;

    CREATE TABLE steps (
        tracing_id      VARCHAR(255) NOT NULL,
        step_index      BIGINT NOT NULL,
        steps_remaining BIGINT NOT NULL,
        start_time      BIGINT NOT NULL,
        error           TEXT,
        PRIMARY KEY (tracing_id, step_index)
    );

COMMIT;
