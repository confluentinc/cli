#!/bin/bash

function cleanup {
  shred --force --remove --zero --iterations=10 deb-secret.gpg rpm-secret.gpg deb-passphrase rpm-passphrase
  rm -rf vendor
}
trap cleanup EXIT

trap "exit 1" ERR

function dry-run {
  if [ "$DRY_RUN" = "true" ]; then
    echo "[DRY_RUN] $1"
  else
    $1
  fi
}

rm -rf deb/ rpm/
mkdir -p deb rpm

aws s3 sync s3://confluent-cli-internal/confluent-cli/deb deb --exclude '*index.html' --exclude '' --exclude '*/'
aws s3 sync s3://confluent-cli-internal/confluent-cli/rpm rpm --exclude '*index.html' --exclude '' --exclude '*/'

aws ecr get-login-password --region us-west-1 | docker login --username AWS --password-stdin 050879227952.dkr.ecr.us-west-1.amazonaws.com

export VAULT_ADDR=https://vault.cireops.gcp.internal.confluent.cloud
vault login -method=oidc -path=okta
vault kv get -field deb_gpg_secret_key v1/ci/kv/cli/release-test > deb-secret.gpg
vault kv get -field deb_gpg_passphrase v1/ci/kv/cli/release-test > deb-passphrase
vault kv get -field rpm_gpg_secret_key v1/ci/kv/cli/release-test > rpm-secret.gpg
vault kv get -field rpm_gpg_passphrase v1/ci/kv/cli/release-test > rpm-passphrase

go mod vendor

# Build linux/amd64
docker build . --file ./docker/Dockerfile_linux_amd64 --tag cli-linux-amd64-builder-image --secret id=deb_gpg_secret_key,src=deb-secret.gpg --secret id=deb_gpg_passphrase,src=deb-passphrase \
  --secret id=rpm_gpg_secret_key,src=rpm-secret.gpg --secret id=rpm_gpg_passphrase,src=rpm-passphrase
docker container create --name cli-linux-amd64-builder cli-linux-amd64-builder-image
docker container cp cli-linux-amd64-builder:/cli/prebuilt/. ./prebuilt/
docker container rm cli-linux-amd64-builder

# Build linux/arm64
architecture=$(uname -m)
if [ "$architecture" == 'x86_64' ]; then
  docker build . --file ./docker/Dockerfile_linux_arm64_from_amd64 --tag cli-linux-arm64-builder-image --secret id=deb_gpg_secret_key,src=deb-secret.gpg --secret id=deb_gpg_passphrase,src=deb-passphrase \
    --secret id=rpm_gpg_secret_key,src=rpm-secret.gpg --secret id=rpm_gpg_passphrase,src=rpm-passphrase
else
  docker build . --file ./docker/Dockerfile_linux_arm64 --tag cli-linux-arm64-builder-image --secret id=deb_gpg_secret_key,src=deb-secret.gpg --secret id=deb_gpg_passphrase,src=deb-passphrase \
    --secret id=rpm_gpg_secret_key,src=rpm-secret.gpg --secret id=rpm_gpg_passphrase,src=rpm-passphrase
fi
docker container create --name cli-linux-arm64-builder cli-linux-arm64-builder-image
docker container cp cli-linux-arm64-builder:/cli/prebuilt/. ./prebuilt/
docker container cp cli-linux-arm64-builder:/cli/deb/. ./deb/
docker container cp cli-linux-arm64-builder:/cli/rpm/. ./rpm/
docker container rm cli-linux-arm64-builder

#TODO: move deb/rpm package uploading to .goreleaser.yml, create new make target to upload the repo metadata & run s3-repo-utils, and remove this block
dry-run "aws s3 sync deb s3://confluent-cli-internal/confluent-cli/deb"
dry-run "aws s3 sync rpm s3://confluent-cli-internal/confluent-cli/rpm"
dry-run "s3-repo-utils -v website index --fake-index --prefix confluent-cli confluent-cli-internal"
