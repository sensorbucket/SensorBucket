CREATE OR REPLACE VIEW enriched_steps_view AS (
    SELECT
        currentstep.*,
        (
            COALESCE(nextstep.start_time, currentstep.start_time) - currentstep.start_time
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
                    -- LIMIT
                    --     10
                ) trace_ids
                LEFT JOIN steps steps ON trace_ids.tracing_id = steps.tracing_id
        ) currentStep
        LEFT JOIN steps nextstep ON currentstep.tracing_id = nextstep.tracing_id
        AND nextstep.step_index = currentstep.step_index + 1
);