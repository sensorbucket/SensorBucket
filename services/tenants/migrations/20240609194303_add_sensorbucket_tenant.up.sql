BEGIN;
    INSERT INTO tenants (id, name, address, zip_code, city, state, logo, created)
    SELECT 1, 'SensorBucket', 'N/A', 'N/A', 'N/A', 1, 'https://raw.githubusercontent.com/sensorbucket/SensorBucket/main/docs/sensorbucket-logo-full.png', NOW()
    WHERE NOT EXISTS (SELECT 1 FROM tenants WHERE id = 1);
    SELECT setval('tenants_id_seq', (SELECT MAX(id) from "tenants"));
COMMIT;
