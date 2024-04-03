-- TestGetPermissions
INSERT INTO tenants (
    id, name, address, zip_code, city, state, parent_tenant_id, created
) VALUES 
    (9, 'Industrial Comp', 'Somestreet 123A', '1234AB', 'Citee', 1, NULL, NOW()),
    (10, 'Acme Inc.', 'Somestreet 123A', '1234AB', 'Citee', 1, 9, NOW()),
    (11, 'Other Acme Inc.', 'Somestreet 123A', '1234AB', 'Citee', 1, NULL, NOW());
    -- 12  does not exist on purpose
    --(12, 'Other Acme Inc.', 'Somestreet 123A', '1234AB', 'Citee', 1, NULL, NOW());

INSERT INTO tenant_members (
    tenant_id, user_id, permissions
) VALUES 
    (9, '67f55001-36f4-4882-8034-63311dcc7523', '{WRITE_API_KEYS}'),
    (10, '67f55001-36f4-4882-8034-63311dcc7523', '{READ_DEVICES,WRITE_DEVICES}');


