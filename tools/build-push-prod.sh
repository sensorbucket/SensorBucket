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
docker build -t ${PROD_CR}/core:${PROD_VER} -f services/core/Dockerfile . 
docker build -t ${PROD_CR}/tracing:${PROD_VER} -f services/tracing/Dockerfile . 
docker build -t ${PROD_CR}/dashboard:${PROD_VER} -f services/dashboard/Dockerfile .
docker build -t ${PROD_CR}/fission-user-workers:${PROD_VER} -f services/fission-user-workers/Dockerfile . 
docker build -t ${PROD_CR}/fission-rmq-connector:${PROD_VER} -f services/fission-rmq-connector/Dockerfile . 

docker push ${PROD_CR}/httpimporter:${PROD_VER}
docker push ${PROD_CR}/core:${PROD_VER}
docker push ${PROD_CR}/tracing:${PROD_VER}
docker push ${PROD_CR}/dashboard:${PROD_VER}
docker push ${PROD_CR}/fission-user-workers:${PROD_VER}
docker push ${PROD_CR}/fission-rmq-connector:${PROD_VER}
