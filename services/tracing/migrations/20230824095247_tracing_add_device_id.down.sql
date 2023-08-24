BEGIN;

   ALTER TABLE steps
   DROP COLUMN device_id;

COMMIT;