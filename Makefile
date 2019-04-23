ALL_SRC         := $(shell find . -name "*.go" | grep -v -e vendor)
GIT_REMOTE_NAME ?= origin
MASTER_BRANCH   ?= master
RELEASE_BRANCH  ?= master

include ./semver.mk

REF := $(shell [ -d .git ] && git rev-parse --short HEAD || echo "none")
DATE := $(shell date -u)
HOSTNAME := $(shell id -u -n)@$(shell hostname -f)

.PHONY: clean
clean:
	rm -rf $(shell pwd)/dist

.PHONY: deps
deps:
	@GO111MODULE=on go get github.com/goreleaser/goreleaser@v0.101.0
	@GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.12.2

build: build-go

ifeq ($(shell uname),Darwin)
GORELEASER_SUFFIX ?= -mac.yml
else
GORELEASER_SUFFIX ?= -linux.yml
endif

show-args:
	@echo "VERSION: $(VERSION)"

.PHONY: build-go
build-go:
	make build-ccloud
	make build-confluent

.PHONY: build-ccloud
build-ccloud:
	@GO111MODULE=on VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --snapshot --rm-dist -f .goreleaser-ccloud$(GORELEASER_SUFFIX)

.PHONY: build-confluent
build-confluent:
	@GO111MODULE=on VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --snapshot --rm-dist -f .goreleaser-confluent$(GORELEASER_SUFFIX)

.PHONY: release
release: get-release-image commit-release tag-release
	make gorelease
	make publish

.PHONY: gorelease
gorelease:
	@GO111MODULE=off go get -u github.com/inconshreveable/mousetrap # dep from cobra -- incompatible with go mod
	@GO111MODULE=on VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --rm-dist -f .goreleaser-ccloud.yml
	@GO111MODULE=on VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --rm-dist -f .goreleaser-confluent.yml

.PHONY: dist-stuff
dist-stuff:
	@# unfortunately goreleaser only supports one archive right now (either tar/zip or binaries): https://github.com/goreleaser/goreleaser/issues/705
	@# we had goreleaser upload binaries (they're uncompressed, so goreleaser's parallel uploads will save more time with binaries than archives)
	@for os in darwin linux windows; do \
		for arch in amd64 386; do \
			if [ "$${os}" = "darwin" ] && [ "$${arch}" = "386" ] ; then \
				continue ; \
			fi; \
			cp LICENSE dist/$(NAME)/$${os}_$${arch}/ ; \
			cp INSTALL.md dist/$(NAME)/$${os}_$${arch}/ ; \
			cd dist/$(NAME)/$${os}_$${arch}/ ; \
			mkdir tmp ; mv LICENSE INSTALL.md $(NAME)* tmp/ ; mv tmp $(NAME) ; \
			suffix="" ; \
			if [ "$${os}" = "windows" ] ; then \
				suffix=zip ; \
				zip -qr ../$(NAME)_$(VERSION)_$${os}_$${arch}.$${suffix} $(NAME) ; \
			else \
				suffix=tar.gz ; \
				tar -czf ../$(NAME)_$(VERSION)_$${os}_$${arch}.$${suffix} $(NAME) ; \
			fi ; \
			cd ../../../ ; \
			cp dist/$(NAME)/$(NAME)_$(VERSION)_$${os}_$${arch}.$${suffix} dist/$(NAME)/$(NAME)_latest_$${os}_$${arch}.$${suffix} ; \
		done ; \
	done

.PHONY: dist-ccloud
dist-ccloud:
	make dist-stuff NAME=ccloud

.PHONY: dist-confluent
dist-confluent:
	make dist-stuff NAME=confluent

.PHONY: dist
dist: dist-ccloud dist-confluent

.PHONY: publish-stuff
publish-stuff:
	aws s3 cp dist/$(NAME)/ s3://confluent.cloud/$(NAME)-cli/archives/$(VERSION:v%=%)/ --recursive --exclude "*" --include "*.tar.gz" --include "*.zip" --include "*_checksums.txt" --exclude "*_latest_*" --acl public-read
	aws s3 cp dist/$(NAME)/ s3://confluent.cloud/$(NAME)-cli/archives/latest/ --recursive --exclude "*" --include "*.tar.gz" --include "*.zip" --include "*_checksums.txt" --exclude "*_$(VERSION)_*" --acl public-read

.PHONY: publish-ccloud
publish-ccloud: dist-ccloud
	make publish-stuff NAME=ccloud

.PHONY: publish-confluent
publish-confluent: dist-confluent
	make publish-stuff NAME=confluent

.PHONY: publish
publish: publish-ccloud publish-confluent

.PHONY: fmt
fmt:
	@gofmt -e -s -l -w $(ALL_SRC)

.PHONY: release-ci
release-ci:
ifeq ($(SEMAPHORE_GIT_BRANCH),master)
	make release
else
	true
endif

cmd/lint/en_US.aff:
	@curl -s "https://chromium.googlesource.com/chromium/deps/hunspell_dictionaries/+/master/en_US.aff?format=TEXT" | base64 -D > $@

cmd/lint/en_US.dic:
	@curl -s "https://chromium.googlesource.com/chromium/deps/hunspell_dictionaries/+/master/en_US.dic?format=TEXT" | base64 -D > $@

.PHONY: lint-cli
lint-cli: cmd/lint/en_US.aff cmd/lint/en_US.dic
	GO111MODULE=on go run cmd/lint/main.go -aff-file $(word 1,$^) -dic-file $(word 2,$^) $(ARGS)

.PHONY: lint-go
lint-go:
	@GO111MODULE=on golangci-lint run

.PHONY: lint
lint: lint-go

.PHONY: coverage
coverage:
      ifdef CI
	@echo "" > coverage.txt
	@for d in $$(go list ./... | grep -v vendor); do \
	  GO111MODULE=on go test -v -race -coverprofile=profile.out -covermode=atomic $$d || exit 2; \
	  if [ -f profile.out ]; then \
	    cat profile.out >> coverage.txt; \
	    rm profile.out; \
	  fi; \
	done
      else
	@GO111MODULE=on go test -race -cover $(TEST_ARGS) $$(go list ./... | grep -v vendor)
      endif

.PHONY: test
test: lint coverage
