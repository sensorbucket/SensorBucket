BEGIN;

alter table api_keys add column permissions VARCHAR[] NOT NULL DEFAULT '{}';

UPDATE api_keys ak
SET permissions = (
    SELECT ARRAY_AGG(akp.permission)
    FROM api_key_permissions akp
    WHERE akp.api_key_id = ak.id
);

drop table api_key_permissions;

COMMIT;
