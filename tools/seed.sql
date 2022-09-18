CREATE DATABASE deviceservice;
CREATE USER deviceservice WITH ENCRYPTED PASSWORD 'deviceservice';
GRANT ALL PRIVILEGES ON DATABASE deviceservice TO deviceservice;
\c deviceservice
CREATE EXTENSION postgis;
