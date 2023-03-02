
BEGIN;

-- Add location columns to devices, 
-- move data from locations table to device columns and remove old columns / table
ALTER TABLE "devices" 
    ADD COLUMN "location" geography,
    ADD COLUMN "location_description" VARCHAR DEFAULT('') NOT NULL;
UPDATE "devices" SET 
    "location" = loc."location",
    "location_description" = loc."name"
FROM "devices" dev LEFT JOIN "locations" loc ON loc.id = dev.location_id;
ALTER TABLE "devices" DROP COLUMN "location_id";
DROP TABLE "locations";

-- Add new fields to sensor
ALTER TABLE "sensors"
    ADD COLUMN "archive_time" INT,
    ADD COLUMN "brand" VARCHAR;

COMMIT;
