BEGIN;

ALTER TABLE pipelines
ADD COLUMN name text;

COMMIT;
