CREATE TABLE "locations" (
  "id" INT GENERATED ALWAYS AS IDENTITY NOT NULL,
  "name" VARCHAR NOT NULL,
  "organisation" VARCHAR NOT NULL,
  "location" geography NOT NULL,

  PRIMARY KEY("id")
);

CREATE TABLE "devices" (
  "id" INT GENERATED ALWAYS AS IDENTITY NOT NULL, 
  "code" VARCHAR NOT NULL,
  "description" VARCHAR NOT NULL DEFAULT(''), 
  "organisation" VARCHAR NOT NULL,
  "configuration" JSON NOT NULL DEFAULT('{}'::json),
  "location_id" INT REFERENCES locations("id"),

  PRIMARY KEY("id")
);

CREATE TABLE "sensors" (
  "id" INT GENERATED ALWAYS AS IDENTITY NOT NULL, 
  "code" VARCHAR NOT NULL,
  "device_id" INT NOT NULL REFERENCES "devices"("id"),
  "description" VARCHAR NOT NULL DEFAULT(''), 
  "external_id" VARCHAR, 
  "configuration" JSON NOT NULL DEFAULT('{}'::json),

  PRIMARY KEY("id")
);

