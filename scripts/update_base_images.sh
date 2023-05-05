#!/bin/bash

scripts/update_Dockerfile_linux_glibc_amd64.sh
scripts/update_Dockerfile_linux_glibc_arm64_from_amd64.sh
scripts/update_Dockerfile_linux_glibc_arm64.sh
echo -e "\nDon't forget to commit your changes to the base Dockerfiles."
