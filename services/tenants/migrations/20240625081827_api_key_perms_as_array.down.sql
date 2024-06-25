INSERT INTO api_key_permissions (api_key_id, permission)
SELECT ak.id, UNNEST(ak.permissions) AS permission
FROM api_keys ak;
