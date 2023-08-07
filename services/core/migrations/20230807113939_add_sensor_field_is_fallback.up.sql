BEGIN;

ALTER TABLE sensors ADD "is_fallback" boolean NOT NULL DEFAULT(false);
UPDATE sensors SET is_fallback = true WHERE sensors.external_id is null;

ALTER TABLE sensors ALTER external_ID SET DEFAULT('');
UPDATE sensors SET external_id = '' WHERE external_id is null;
ALTER TABLE sensors ALTER external_id SET NOT NULL;

COMMIT;
