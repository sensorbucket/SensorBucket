#!/bin/sh
set -o errexit

if [ ! -v PROD_CR ] || [ -z "$PROD_CR" ]
then
    echo "PROD_CR not set"
    exit 1
fi

if [ ! -v PROD_VER ] || [ -z "$PROD_VER" ]
then
    echo "PROD_VER not set"
    exit 1
fi

docker build -t ${PROD_CR}/httpimporter:${PROD_VER} -f services/httpimporter/Dockerfile . 
docker build -t ${PROD_CR}/device:${PROD_VER} -f services/device/Dockerfile . 
docker build -t ${PROD_CR}/measurements:${PROD_VER} -f services/measurements/Dockerfile . 
docker build -t ${PROD_CR}/pipeline:${PROD_VER} -f services/pipeline/Dockerfile . 
docker build -t ${PROD_CR}/dashboard:${PROD_VER} services/dashboard/  

docker build -t ${PROD_CR}/worker-the-things-network:${PROD_VER} -f workers/the-things-network/Dockerfile . 
docker build -t ${PROD_CR}/worker-multiflexmeter-groundwater:${PROD_VER} -f workers/multiflexmeter-groundwater-level/Dockerfile . 
docker build -t ${PROD_CR}/worker-multiflexmeter-particulatematter:${PROD_VER} -f workers/multiflexmeter-particulatematter/Dockerfile . 
docker build -t ${PROD_CR}/worker-pzld-sensorbox:${PROD_VER} -f workers/pzld-sensorbox/Dockerfile . 


docker push ${PROD_CR}/httpimporter:${PROD_VER} 
docker push ${PROD_CR}/device:${PROD_VER} 
docker push ${PROD_CR}/measurements:${PROD_VER} 
docker push ${PROD_CR}/pipeline:${PROD_VER} 
docker push ${PROD_CR}/dashboard:${PROD_VER}
docker push ${PROD_CR}/worker-the-things-network:${PROD_VER} 
docker push ${PROD_CR}/worker-multiflexmeter-groundwater:${PROD_VER} 
docker push ${PROD_CR}/worker-multiflexmeter-particulatematter:${PROD_VER} 
docker push ${PROD_CR}/worker-pzld-sensorbox:${PROD_VER} 
