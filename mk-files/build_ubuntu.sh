#!/bin/bash

cp ~/.netrc .
docker build . -f Dockerfile_ubuntu -t cli-ubuntu-builder-image
docker container create --name cli-ubuntu-builder cli-ubuntu-builder-image
docker container cp cli-ubuntu-builder:/go/src/github.com/confluentinc/cli/dist/. ./dist/
docker container rm cli-ubuntu-builder