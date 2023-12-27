#!/bin/bash

function cleanup {
  shred --force --remove --zero --iterations=10 CLIEVCodeSigningCertificate2.pfx
  rm -rf vendor
}
trap cleanup EXIT

trap "exit 1" ERR

az login
az keyvault secret download --file CLIEVCodeSigningCertificate2.pfx --name CLIEVCodeSigningCertificate2 --subscription cc-prod --vault-name CLICodeSigningKeyVault --encoding base64

go mod vendor

# Build windows/amd64
docker build . --file ./docker/Dockerfile_windows_amd64 --tag cli-windows-amd64-builder-image
docker container create --name cli-windows-amd64-builder cli-windows-amd64-builder-image
docker container cp cli-windows-amd64-builder:/cli/prebuilt/. ./prebuilt/
docker container rm cli-windows-amd64-builder
