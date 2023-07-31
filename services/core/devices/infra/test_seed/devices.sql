INSERT INTO public.devices (code, description, organisation, properties, location_description, state) 
VALUES 
  ('D1', 'Device 1 description', 'Org1', '{}', 'Location1', 1),
  ('D2', 'Device 2 description', 'Org2', '{}', 'Location2', 2),
  ('D3', 'Device 3 description', 'Org3', '{}', 'Location3', 3);

INSERT INTO public.sensors (code, device_id, description, properties, brand, external_id) 
VALUES 
  ('S1', 1, 'Sensor 1 description', '{}', 'Brand1', '1'),
  ('S2', 1, 'Sensor 2 description', '{}', 'Brand2', '2'),
  ('S3', 2, 'Sensor 3 description', '{}', 'Brand3', '3'),
  ('S4', 2, 'Sensor 4 description', '{}', 'Brand4', '4'),
  ('S5', 3, 'Sensor 5 description', '{}', 'Brand5', '5'),
  ('S6', 3, 'Sensor 6 description', '{}', 'Brand6', '6');

