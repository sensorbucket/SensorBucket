BEGIN;

  DROP TABLE archived_ingress_dtos CASCADE;
  DROP TABLE steps CASCADE;

  CREATE TABLE trace_ingress (
    id UUID NOT NULL,
    tenant_id BIGINT NOT NULL,
    pipeline_id UUID,
    archived_at TIMESTAMPTZ,
    payload BYTEA
  );
  CREATE INDEX idx_trace_ingress ON trace_ingress(tenant_id, archived_at DESC);

  CREATE TABLE traces (
    id UUID NOT NULL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    pipeline_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    error TEXT,
    error_at TIMESTAMPTZ
  );
  CREATE INDEX idx_trace_pipeline_id ON traces(pipeline_id);
  CREATE INDEX idx_trace_has_error ON traces(error) WHERE error IS NOT NULL;

  CREATE TABLE trace_steps (
    tracing_id UUID NOT NULL,
    worker_id TEXT NOT NULL,
    queue_time TIMESTAMPTZ NOT NULL,
    device_id BIGINT NOT NULL DEFAULT 0
  );
  CREATE INDEX idx_trace_steps_pagination ON trace_steps(tracing_id, queue_time DESC);

COMMIT;
