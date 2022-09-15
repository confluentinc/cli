#!/bin/bash

go mod vendor
docker build . -f ./dockerfiles/Dockerfile_linux_glibc_arm64 -t cli-linux-glibc-arm64-builder-image
docker container create --name cli-linux-glibc-arm64-builder cli-linux-glibc-arm64-builder-image
docker container cp cli-linux-glibc-arm64-builder:/go/src/github.com/confluentinc/cli/dist/. ./dist/
docker container rm cli-linux-glibc-arm64-builder
rm -rf vendor