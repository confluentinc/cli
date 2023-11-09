#!/bin/bash

function cleanup {
  shred --force --remove --zero --iterations=10 secret.gpg passphrase
  rm -rf vendor
}
trap cleanup EXIT

function dry-run {
  if [ "$DRY_RUN" = "true" ]; then
    echo "[DRY_RUN] $1"
  else
    $1
  fi
}

rm -rf deb/ rpm/
mkdir -p deb rpm

# aws s3 sync s3://confluent.cloud.internal/deb deb
aws s3 sync s3://confluent.cloud.internal/rpm rpm --exclude '*index.html' --exclude '' --exclude '*/'

aws ecr get-login-password --region us-west-1 | docker login --username AWS --password-stdin 050879227952.dkr.ecr.us-west-1.amazonaws.com

export VAULT_ADDR=https://vault.cireops.gcp.internal.confluent.cloud
vault login -method=oidc -path=okta
vault kv get -field gpg_secret_key v1/ci/kv/cli/release-test > secret.gpg
vault kv get -field gpg_passphrase v1/ci/kv/cli/release-test > passphrase

go mod vendor

# Build linux/amd64
docker build . --file ./docker/Dockerfile_linux_amd64 --tag cli-linux-amd64-builder-image --secret id=gpg_secret_key,src=secret.gpg --secret id=gpg_passphrase,src=passphrase
docker container create --name cli-linux-amd64-builder cli-linux-amd64-builder-image
docker container cp cli-linux-amd64-builder:/cli/prebuilt/. ./prebuilt/
docker container rm cli-linux-amd64-builder

# Build linux/arm64
architecture=$(uname -m)
if [ "$architecture" == 'x86_64' ]; then
  docker build . --file ./docker/Dockerfile_linux_arm64_from_amd64 --tag cli-linux-arm64-builder-image --secret id=gpg_secret_key,src=secret.gpg --secret id=gpg_passphrase,src=passphrase
else
  docker build . --file ./docker/Dockerfile_linux_arm64 --tag cli-linux-arm64-builder-image --secret id=gpg_secret_key,src=secret.gpg --secret id=gpg_passphrase,src=passphrase
fi
docker container create --name cli-linux-arm64-builder cli-linux-arm64-builder-image
docker container cp cli-linux-arm64-builder:/cli/prebuilt/. ./prebuilt/
docker container cp cli-linux-arm64-builder:/cli/rpm/. ./rpm/
docker container rm cli-linux-arm64-builder

dry-run "aws s3 sync rpm s3://confluent.cloud.internal/rpm"
dry-run "s3-repo-utils -v website index --fake-index --prefix "rpm" confluent.cloud.internal"
