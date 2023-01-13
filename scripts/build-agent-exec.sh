#!/bin/bash

LINUX_OS="linux"
LINUX_ARCH="arm64"

BUILD_OS=$(go env GOOS)
BUILD_ARCH=$(go env GOARCH)

if [ $# -gt 0 ]
    then
        BUILD_OS="$1"
        echo "   Build OS: ${BUILD_OS}"
        if [ $# -gt 1 ]
            then
            BUILD_ARCH="$2"
            echo "   Build ARCH: ${BUILD_ARCH}"
        fi
fi
echo "   Build ENV: ${BUILD_OS}/${BUILD_ARCH}"

BASEDIR=$(dirname "$0")
cd ${BASEDIR}/../cmd/cli

echo "building executable for ${BUILD_OS}/${BUILD_ARCH}"
GOOS="${BUILD_OS}" GOARCH="${BUILD_ARCH}" go build -o build/"${BUILD_OS}_${BUILD_ARCH}"/ comffstore.go