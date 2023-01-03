CREATE TABLE "pipelines" (
  "id" UUID NOT NULL PRIMARY KEY,
  "description" VARCHAR DEFAULT(''),
  "status" VARCHAR NOT NULL,
  "last_status_change" TIMESTAMPZ NOT NULL
);

CREATE TABLE "pipeline_steps" (
  "pipeline_id" UUID NOT NULL REFERENCES "pipelines"("id"),
  "pipeline_step" INT NOT NULL,
  "image" VARCHAR NOT NULL,

  UNIQUE("pipeline_id", "pipeline_step")
);
