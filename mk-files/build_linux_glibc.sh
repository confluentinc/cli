#!/bin/bash

cd ..
cp ~/.netrc .
docker build . -f ./dockerfiles/Dockerfile_linux_glibc -t cli-linux-glibc-builder-image
docker container create --name cli-linux-glibc-builder cli-linux-glibc-builder-image
docker container cp cli-linux-glibc-builder:/go/src/github.com/confluentinc/cli/dist/. ./dist/
docker container rm cli-linux-glibc-builder