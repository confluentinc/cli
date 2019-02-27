# Maintainer: Travis Thompson <trthomps@confluent.io>
pkgname=ccloud
pkgver=0.25.1
pkgrel=1
pkgdesc="Confluent Cloud CLI"
arch=('x86_64')
url="https://github.com/confluentinc/cli"
license=('custom')
makedepends=('go>=1.11.0')
source=("git+https://github.com/confluentinc/cli.git")
md5sums=('SKIP')

pkgver() {
  make -C cli show-version | grep -E '^clean version:' | awk '{print $3}'
}

prepare() {
  make -C cli deps
}

build() {
  GOPATH=$(go env GOPATH) make -C cli build-go
}

package() {
  find cli/dist/linux_amd64/ -type f -exec install -Dm 755 "{}" -t "$pkgdir/usr/bin" \;
}
