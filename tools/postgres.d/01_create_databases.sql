-- Create Device Service database
CREATE DATABASE deviceservice;
CREATE USER deviceservice WITH ENCRYPTED PASSWORD 'deviceservice';
GRANT ALL PRIVILEGES ON DATABASE deviceservice TO deviceservice;
\c deviceservice
CREATE EXTENSION postgis;

-- Create Measurements Service database
CREATE DATABASE measurementservice;
CREATE USER measurementservice WITH ENCRYPTED PASSWORD 'measurementservice';
GRANT ALL PRIVILEGES ON DATABASE measurementservice TO measurementservice;
\c measurementservice
CREATE EXTENSION timescaledb;
CREATE EXTENSION postgis;

-- Create Pipeline Service database
CREATE DATABASE pipelineservice;
CREATE USER pipelineservice WITH ENCRYPTED PASSWORD 'pipelineservice';
GRANT ALL PRIVILEGES ON DATABASE pipelineservice TO pipelineservice;
