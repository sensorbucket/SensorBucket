#!/bin/sh
set -o errexit

docker build -t localhost:5001/httpimporter:latest -f services/httpimporter/Dockerfile .
docker build -t localhost:5001/device:latest -f services/device/Dockerfile .
docker build -t localhost:5001/measurements:latest -f services/measurements/Dockerfile .
docker build -t localhost:5001/pipeline:latest -f services/pipeline/Dockerfile .
docker build -t localhost:5001/dashboard:latest services/dashboard/ 

docker build -t localhost:5001/worker-the-things-network:latest -f workers/the-things-network/Dockerfile .
docker build -t localhost:5001/worker-multiflexmeter-groundwater:latest -f workers/multiflexmeter-groundwater-level/Dockerfile .
docker build -t localhost:5001/worker-multiflexmeter-particulatematter:latest -f workers/multiflexmeter-particulatematter/Dockerfile .
