#!/bin/bash
set -e

# This script is designed to live in a Confluent Platform tarball distribution.
# It depends on the CLI binaries for different OS/ARCH combinations living at
#   ../libexec/cli/${OS}_${ARCH}/confluent
#

is_supported_platform() {
  platform=$1
  found=1
  case "$platform" in
    linux/amd64) found=0 ;;
    linux/386) found=0 ;;
    darwin/amd64) found=0 ;;
    darwin/386) found=0 ;;
    windows/amd64) found=0 ;;
    windows/386) found=0 ;;
  esac
  case "$platform" in
    darwin/386) found=1 ;;
  esac
  return $found
}
check_platform() {
  if is_supported_platform "$PLATFORM"; then
    # optional logging goes here
    true
  else
    log_crit "platform $PLATFORM is not supported.  Make sure this script is up-to-date and file request at https://github.com/${PREFIX}/issues/new"
    exit 1
  fi
}
adjust_os() {
  # adjust archive name based on OS
  case ${OS} in
    386) OS=i386 ;;
    amd64) OS=x86_64 ;;
    darwin) OS=darwin ;;
    linux) OS=linux ;;
    windows) OS=windows ;;
  esac
  true
}
check_executable() {
  if [[ -f "${EXECUTABLE}" ]] ; then
    # optional logging goes here
    true
  else
    log_crit "executable $EXECUTABLE does not exist.  Make sure this script is up-to-date and resides at bin/confluent in the Confluent Platform release directory."
    exit 1
  fi
}
init_config() {
  mkdir -p ${HOME}/.confluent
  if [[ ! -f "${HOME}/.confluent/config.json" ]] ; then
    echo '{"disable_updates": true}' > "${HOME}/.confluent/config.json"
  fi
}

cat /dev/null <<EOF
------------------------------------------------------------------------
https://github.com/client9/shlib - portable posix shell functions
Public domain - http://unlicense.org
https://github.com/client9/shlib/blob/master/LICENSE.md
but credit (and pull requests) appreciated.
------------------------------------------------------------------------
EOF
echoerr() {
  echo "$@" 1>&2
}
log_prefix() {
  echo "$0"
}
_logp=6
log_set_priority() {
  _logp="$1"
}
log_priority() {
  if test -z "$1"; then
    echo "$_logp"
    return
  fi
  [ "$1" -le "$_logp" ]
}
log_tag() {
  case $1 in
    0) echo "emerg" ;;
    1) echo "alert" ;;
    2) echo "crit" ;;
    3) echo "err" ;;
    4) echo "warning" ;;
    5) echo "notice" ;;
    6) echo "info" ;;
    7) echo "debug" ;;
    *) echo "$1" ;;
  esac
}
log_debug() {
  log_priority 7 || return 0
  echoerr "$(log_prefix)" "$(log_tag 7)" "$@"
}
log_info() {
  log_priority 6 || return 0
  echoerr "$(log_prefix)" "$(log_tag 6)" "$@"
}
log_err() {
  log_priority 3 || return 0
  echoerr "$(log_prefix)" "$(log_tag 3)" "$@"
}
log_crit() {
  log_priority 2 || return 0
  echoerr "$(log_prefix)" "$(log_tag 2)" "$@"
}
uname_os() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
    msys*) os="windows" ;;
    mingw*) os="windows" ;;
  esac
  echo "$os"
}
uname_arch() {
  arch=$(uname -m)
  case $arch in
    x86_64) arch="amd64" ;;
    x86) arch="386" ;;
    i686) arch="386" ;;
    i386) arch="386" ;;
    aarch64) arch="arm64" ;;
    armv5*) arch="armv5" ;;
    armv6*) arch="armv6" ;;
    armv7*) arch="armv7" ;;
  esac
  echo ${arch}
}
uname_os_check() {
  os=$(uname_os)
  case "$os" in
    darwin) return 0 ;;
    dragonfly) return 0 ;;
    freebsd) return 0 ;;
    linux) return 0 ;;
    android) return 0 ;;
    nacl) return 0 ;;
    netbsd) return 0 ;;
    openbsd) return 0 ;;
    plan9) return 0 ;;
    solaris) return 0 ;;
    windows) return 0 ;;
  esac
  log_crit "uname_os_check '$(uname -s)' got converted to '$os' which is not a GOOS value. Please file bug at https://github.com/client9/shlib"
  return 1
}
uname_arch_check() {
  arch=$(uname_arch)
  case "$arch" in
    386) return 0 ;;
    amd64) return 0 ;;
    arm64) return 0 ;;
    armv5) return 0 ;;
    armv6) return 0 ;;
    armv7) return 0 ;;
    ppc64) return 0 ;;
    ppc64le) return 0 ;;
    mips) return 0 ;;
    mipsle) return 0 ;;
    mips64) return 0 ;;
    mips64le) return 0 ;;
    s390x) return 0 ;;
    amd64p32) return 0 ;;
  esac
  log_crit "uname_arch_check '$(uname -m)' got converted to '$arch' which is not a GOARCH value.  Please file bug report at https://github.com/client9/shlib"
  return 1
}
cat /dev/null <<EOF
------------------------------------------------------------------------
End of functions from https://github.com/client9/shlib
------------------------------------------------------------------------
EOF

OWNER=confluentinc
REPO="cli"
BINARY=confluent
OS=$(uname_os)
ARCH=$(uname_arch)
PREFIX="${OWNER}/${REPO}"
EXE_PATH="${BASH_SOURCE%/*}/../libexec/cli"
PLATFORM="${OS}/${ARCH}"

# use in logging routines
log_prefix() {
	echo "${PREFIX}"
}

uname_os_check
uname_arch_check

check_platform

adjust_os

EXECUTABLE="${EXE_PATH}/${OS}_${ARCH}/${BINARY}"
if [[ "$OS" = "windows" ]]; then
  EXECUTABLE="${EXECUTABLE}.exe"
fi

check_executable
init_config

# call the underlying executable
${EXECUTABLE} $@
