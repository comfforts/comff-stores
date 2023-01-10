#!/bin/bash

CURR_SHA=$(git rev-parse --short HEAD)

BASEDIR=$(dirname "$0")
cd ${BASEDIR}/../
echo "running github.com/comfforts/comff-stores:${CURR_SHA}"

docker run --rm -p 50051:50051 --name comff-stores github.com/comfforts/comff-stores:${CURR_SHA}