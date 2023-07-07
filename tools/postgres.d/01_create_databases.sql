CREATE DATABASE sensorbucket;
CREATE USER sensorbucket WITH ENCRYPTED PASSWORD 'sensorbucket';
GRANT ALL PRIVILEGES ON DATABASE sensorbucket TO sensorbucket;
\c sensorbucket
CREATE EXTENSION postgis;
CREATE EXTENSION timescaledb;
