#!/bin/bash

CURR_SHA=$(git rev-parse --short HEAD)
CURR_BRANCH=$(git rev-parse --abbrev-ref HEAD)
BASEDIR=$(dirname "$0")
echo $BASEDIR

cd ${BASEDIR}/../
echo "building docker image for rev: ${CURR_SHA} for branch: ${CURR_BRANCH}"
docker build --progress=plain --build-arg BUILD_OS=linux --build-arg BUILD_ARCH=arm64 -t comff-stores:${CURR_SHA} -f docker/Dockerfile .