INSERT INTO public.devices (code, description, tenant_id, properties, location_description, state) 
VALUES 
  ('D1', 'Device 1 description', 10, '{}', 'Location1', 1),
  ('D2', 'Device 2 description', 10, '{}', 'Location2', 2),
  ('D3', 'Device 3 description', 10, '{}', 'Location3', 3);

INSERT INTO public.sensors (code, device_id, description, properties, brand, external_id, tenant_id) 
VALUES 
  ('S1', 1, 'Sensor 1 description', '{}', 'Brand1', '1', 10),
  ('S2', 1, 'Sensor 2 description', '{}', 'Brand2', '2', 10),
  ('S3', 2, 'Sensor 3 description', '{}', 'Brand3', '3', 10),
  ('S4', 2, 'Sensor 4 description', '{}', 'Brand4', '4', 10),
  ('S5', 3, 'Sensor 5 description', '{}', 'Brand5', '5', 10),
  ('S6', 3, 'Sensor 6 description', '{}', 'Brand6', '6', 10);

