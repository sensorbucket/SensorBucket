CREATE TABLE locations (
   id serial PRIMARY KEY,
   name VARCHAR ( 50 ) UNIQUE NOT NULL,
   lat float8 DEFAULT 0,
   lng float8 DEFAULT 0
);

CREATE TABLE thing_locations (
   urn VARCHAR ( 150 ) NOT NULL,
   location_id INT NOT NULL,
   PRIMARY KEY (urn),
   FOREIGN KEY (location_id)
      REFERENCES locations (id)
);