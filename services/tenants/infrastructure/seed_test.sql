INSERT INTO tenants (
    id, name, address, zip_code, city, state, created
) VALUES 
    (10, 'Acme Inc.', 'Somestreet 123A', '1234AB', 'Citee', 1, NOW()),
    (11, 'Other Acme Inc.', 'Somestreet 123A', '1234AB', 'Citee', 1, NOW());

INSERT INTO tenant_members (
    tenant_id, user_id, permissions
) VALUES ( 
    10, '67f55001-36f4-4882-8034-63311dcc7523', '{READ_DEVICES,WRITE_DEVICES}' 
);


