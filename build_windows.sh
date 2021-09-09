#!/bin/bash

cp ~/.netrc .
docker build . -f Dockerfile_windows -t cli-windows-builder-image
docker container create --name cli-windows-builder cli-windows-builder-image
docker container cp cli-windows-builder:/go/src/github.com/confluentinc/cli/dist/. ./dist/
docker container rm cli-windows-builder
