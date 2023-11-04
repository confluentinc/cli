#!/bin/bash

rm -rf deb/ rpm/
mkdir -p deb rpm

# aws s3 sync s3://confluent.cloud.internal/deb deb
aws s3 sync s3://confluent.cloud.internal/rpm rpm

aws ecr get-login-password --region us-west-1 | docker login --username AWS --password-stdin 050879227952.dkr.ecr.us-west-1.amazonaws.com

go mod vendor

# Build linux/amd64
docker build . --file ./docker/Dockerfile_linux_amd64 --tag cli-linux-amd64-builder-image --secret id=gpg_private_key,src=secret.gpg --secret id=gpg_passphrase,src=passphrase
docker container create --name cli-linux-amd64-builder cli-linux-amd64-builder-image
docker container cp cli-linux-amd64-builder:/cli/prebuilt/. ./prebuilt/
docker container rm cli-linux-amd64-builder

# Build linux/arm64
architecture=$(uname -m)
if [ "$architecture" == 'x86_64' ]; then
  docker build . --file ./docker/Dockerfile_linux_arm64_from_amd64 --tag cli-linux-arm64-builder-image --secret id=gpg_private_key,src=secret.gpg --secret id=gpg_passphrase,src=passphrase
else
  docker build . --file ./docker/Dockerfile_linux_arm64 --tag cli-linux-arm64-builder-image --secret id=gpg_private_key,src=secret.gpg --secret id=gpg_passphrase,src=passphrase
fi
docker container create --name cli-linux-arm64-builder cli-linux-arm64-builder-image
docker container cp cli-linux-arm64-builder:/cli/prebuilt/. ./prebuilt/
docker container cp cli-linux-arm64-builder:/cli/rpm/. ./rpm/
docker container rm cli-linux-arm64-builder

rm -rf vendor