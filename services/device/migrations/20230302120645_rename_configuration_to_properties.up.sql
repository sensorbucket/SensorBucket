BEGIN;

ALTER TABLE "devices" RENAME COLUMN "configuration" TO "properties";
ALTER TABLE "sensors" RENAME COLUMN "configuration" TO "properties";

COMMIT;
