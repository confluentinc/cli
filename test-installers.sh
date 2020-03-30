#!/usr/bin/env bash
set -e

# values set in test loop below
TEST_OS=
TEST_ARCH=

# mock out `uname`
uname() {
  if [[ "$1" = "-s" ]]; then
    echo ${TEST_OS}
  elif [[ "$1" = "-m" ]]; then
    echo ${TEST_ARCH}
  fi
}

TEST=true . install-ccloud.sh  # commands are identical for ccloud and confluent, enforced by linter

# useful reference for valid uname system/machine pairs: https://en.wikipedia.org/wiki/Uname
for pair in Darwin,x86_64,darwin,amd64 Linux,x86_64,linux,amd64 Linux,i686,linux,386 CYGWIN_NT-10.0,386,windows,386 CYGWIN_NT-10.0,x86_64,windows,amd64; do
  # tip from https://stackoverflow.com/a/36393986/337735
  IFS=',' read TEST_OS TEST_ARCH EXPECT_OS EXPECT_ARCH <<< "${pair}"

  [[ "$(uname_os)" = "${EXPECT_OS}" ]] || ( echo "${TEST_OS}: got uname_os $(uname_os), want ${EXPECT_OS}" && exit 1 )
  [[ "$(uname_arch)" = "${EXPECT_ARCH}" ]] || ( echo "${TEST_OS}: got uname_arch $(uname_arch), want ${EXPECT_ARCH}" && exit 1 )

  uname_os_check
  uname_arch_check
done

for cliBinary in ccloud confluent; do
  TEST_OS=$(go env GOOS)
  TEST_ARCH=$(go env GOARCH)
  output=$(./install-${cliBinary}.sh 2>&1)
  tmpdir=$(echo "${output}" | sed -n 's/.*licenses located in \(.*\)/\1/p')

  ls "${tmpdir}" | grep -q "LICENSE" || ( echo "License file not found" && exit 1 )
  [[ "$(ls "${tmpdir}/legal/licenses" | wc -l)" -ge 20 ]] || ( echo "Appears to be missing some licenses; found less than 20 in the tmp dir" && exit 1 )

  if [[ "${cliBinary}" = "ccloud" ]]; then
    cliName="Confluent Cloud"
  else
    cliName="Confluent Platform"
  fi
  ./bin/${cliBinary} -h 2>&1 >/dev/null | grep -q "Manage your ${cliName}." || ( echo "Unable to execute installed ${cliBinary} CLI" && exit 1 )

  #rm -rf ${tmpdir}  # too scary for now
done

echo "All tests passed!"
