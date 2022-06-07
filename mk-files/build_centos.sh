#!/bin/bash

cd ..
cp ~/.netrc .
docker build . -f Dockerfile_centos -t cli-centos-builder-image
docker container create --name cli-centos-builder cli-centos-builder-image
docker container cp cli-centos-builder:/go/src/github.com/confluentinc/cli/dist/. ./dist/
docker container rm cli-centos-builder