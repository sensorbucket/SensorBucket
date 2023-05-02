#!/bin/sh
set -o errexit

docker push localhost:5001/httpimporter:latest
docker push localhost:5001/measurements:latest
docker push localhost:5001/pipeline:latest
docker push localhost:5001/dashboard:latest

docker push localhost:5001/worker-the-things-network:latest
docker push localhost:5001/worker-multiflexmeter-groundwater:latest
docker push localhost:5001/worker-multiflexmeter-particulatematter:latest
