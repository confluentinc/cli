#!/bin/bash

aws ecr get-login-password --region us-west-1 | docker login --username AWS --password-stdin 050879227952.dkr.ecr.us-west-1.amazonaws.com

go mod vendor

# Build linux/amd64
docker build . -f ./dockerfiles/Dockerfile_linux_glibc -t cli-linux-glibc-builder-image
docker container create --name cli-linux-glibc-builder cli-linux-glibc-builder-image
docker container cp cli-linux-glibc-builder:/go/src/github.com/confluentinc/cli/dist/. ./dist/
docker container rm cli-linux-glibc-builder

# Build linux/arm64
architecture=$(uname -m)
if [ "$architecture" == 'x86_64' ]; then
  docker build . -f ./dockerfiles/Dockerfile_linux_glibc_arm64_from_amd64 -t cli-linux-glibc-arm64-builder-image
else
  docker build . -f ./dockerfiles/Dockerfile_linux_glibc_arm64 -t cli-linux-glibc-arm64-builder-image
fi
docker container create --name cli-linux-glibc-arm64-builder cli-linux-glibc-arm64-builder-image
docker container cp cli-linux-glibc-arm64-builder:/go/src/github.com/confluentinc/cli/dist/. ./dist/
docker container rm cli-linux-glibc-arm64-builder
rm -rf vendor