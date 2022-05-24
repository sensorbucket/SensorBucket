CREATE TABLE locations (
   id serial PRIMARY KEY,
   name VARCHAR UNIQUE NOT NULL,
   latitude float8 DEFAULT 0,
   longitude float8 DEFAULT 0
);

CREATE TABLE thing_locations (
   thing_urn VARCHAR NOT NULL,
   location_id INT NOT NULL,
   PRIMARY KEY (thing_urn),
   FOREIGN KEY (location_id)
      REFERENCES locations (id)
);