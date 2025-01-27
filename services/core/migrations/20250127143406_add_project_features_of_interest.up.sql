BEGIN;

  CREATE TABLE projects (
    id BIGINT NOT NULL GENERATED ALWAYS AS IDENTITY,
    name VARCHAR NOT NULL,
    description VARCHAR NOT NULL DEFAULT '',
    tenant_id BIGINT NOT NULL
  );

  CREATE TABLE feature_of_interest (
    id BIGINT NOT NULL GENERATED ALWAYS AS IDENTITY,
    name VARCHAR NOT NULL,
    description VARCHAR NOT NULL DEFAULT '',
    tenant_id BIGINT NOT NULL
  );

  CREATE TABLE project_feature_of_interest (
    project_id BIGINT REFERENCES projects(id),
    feature_of_interest_id BIGINT REFERENCES feature_of_interest(id),
    interested_observation_types []VARCHAR NOT NULL,

    PRIMARY KEY(project_id)
  );

  CREATE TABLE feature_of_interest_datastream (
    feature_of_interest_id BIGINT REFERENCES feature_of_interest(id),
    id UUID NOT NULL REFERENCES datastreams(id),
    bound_at TIMESTAMPTZ NOT NULL,
    unbound_at TIMESTAMPTZ,

    PRIMARY KEY(feature_of_interest_id)
  );

  ALTER TABLE measurements
     ADD COLUMN feature_of_interest_id BIGINT;
  CREATE INDEX measurements_query_by_foi (measurements_timestamp DESC, feature_of_interest_id);

COMMIT;
