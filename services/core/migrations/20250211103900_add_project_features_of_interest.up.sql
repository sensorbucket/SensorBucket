CREATE TABLE projects (
  id BIGINT NOT NULL GENERATED ALWAYS AS IDENTITY,
  name VARCHAR NOT NULL,
  description VARCHAR NOT NULL DEFAULT '',
  tenant_id BIGINT NOT NULL,

  PRIMARY KEY(id)
);

CREATE TABLE features_of_interest (
  id BIGINT NOT NULL GENERATED ALWAYS AS IDENTITY,
  name VARCHAR NOT NULL,
  description VARCHAR NOT NULL DEFAULT '',
  encoding_type VARCHAR, -- Oneof:
  feature GEOMETRY,
  properties JSONB NOT NULL DEFAULT '{}'::json,
  tenant_id BIGINT NOT NULL,

  PRIMARY KEY(id)
);

CREATE TABLE project_feature_of_interest (
  project_id BIGINT NOT NULL REFERENCES projects(id),
  feature_of_interest_id BIGINT NOT NULL REFERENCES features_of_interest(id),
  interested_observation_types VARCHAR[] NOT NULL
);

CREATE INDEX project_feature_of_interest_idx ON project_feature_of_interest(project_id);

ALTER TABLE sensors ADD COLUMN feature_of_interest_id BIGINT REFERENCES features_of_interest(id);
ALTER TABLE measurements 
  ADD COLUMN feature_of_interest_id BIGINT,
  ADD COLUMN feature_of_interest_name TEXT,
  ADD COLUMN feature_of_interest_description TEXT,
  ADD COLUMN feature_of_interest_encoding_type TEXT,
  ADD COLUMN feature_of_interest_feature GEOMETRY,
  ADD COLUMN feature_of_interest_properties JSONB;
CREATE INDEX measurements_query_by_foi ON measurements(feature_of_interest_id, datastream_observed_property, measurement_timestamp DESC);
