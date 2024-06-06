#!/bin/bash

function cleanup {
  shred --force --remove --zero --iterations=10 deb-secret.gpg rpm-secret.gpg deb-passphrase rpm-passphrase
}
trap cleanup EXIT

trap "exit 1" ERR

rm -rf prebuilt/ deb/ rpm/
mkdir prebuilt/ deb/ rpm/

aws s3 cp s3://confluent-cli-release/confluent-cli/deb deb --recursive --exclude '*index.html' --exclude '' --exclude '*/' --exclude '*.deb'
aws s3 cp s3://confluent-cli-release/confluent-cli/rpm rpm --recursive --exclude '*index.html' --exclude '' --exclude '*/' --exclude '*.rpm'

gh release download $VERSION -p '*.deb' -p '*.rpm' --dir prebuilt/

export VAULT_ADDR=https://vault.cireops.gcp.internal.confluent.cloud
vault login -method=oidc -path=okta
vault kv get -field deb_gpg_secret_key v1/ci/kv/cli/release > deb-secret.gpg
vault kv get -field deb_gpg_passphrase v1/ci/kv/cli/release > deb-passphrase
vault kv get -field rpm_gpg_secret_key v1/ci/kv/cli/release > rpm-secret.gpg
vault kv get -field rpm_gpg_passphrase v1/ci/kv/cli/release > rpm-passphrase

# Build APT/YUM repos
aws ecr get-login-password --region us-west-1 | docker login --username AWS --password-stdin 050879227952.dkr.ecr.us-west-1.amazonaws.com

docker build . --file ./docker/Dockerfile_linux_repos --tag cli-linux-repo-update-image --secret id=deb_gpg_secret_key,src=deb-secret.gpg --secret id=deb_gpg_passphrase,src=deb-passphrase \
  --secret id=rpm_gpg_secret_key,src=rpm-secret.gpg --secret id=rpm_gpg_passphrase,src=rpm-passphrase
docker container create --name cli-linux-repo-update cli-linux-repo-update-image
docker container cp cli-linux-repo-update:/cli/deb/. ./deb/
docker container cp cli-linux-repo-update:/cli/rpm/. ./rpm/
docker container rm cli-linux-repo-update
