#!/bin/bash

cp ~/.netrc .
docker build . -f Dockerfile_alpine -t cli-alpine-builder-image
docker container create --name cli-alpine-builder cli-alpine-builder-image
docker container cp cli-alpine-builder:/go/src/github.com/confluentinc/cli/dist/. ./dist/
docker container rm cli-alpine-builder
