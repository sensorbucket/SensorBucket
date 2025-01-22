DROP INDEX IF EXISTS idx_trace_pipeline_id;
DROP INDEX IF EXISTS idx_trace_has_error;
DROP INDEX IF EXISTS idx_trace_steps_pagination;

CREATE INDEX idx_trace_latest_for_pipeline ON traces(tenant_id, pipeline_id, created_at DESC);
CREATE INDEX idx_trace_steps_latest_for_pipeline ON trace_steps(tracing_id, queue_time DESC);
