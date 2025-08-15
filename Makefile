SHELL := /bin/bash
GORELEASER_VERSION := v1.21.2

# Compile natively based on the current system
.PHONY: build
build:
ifneq "" "$(findstring NT,$(shell uname))" # windows
	CC=gcc CXX=g++ $(MAKE) cli-builder
else ifneq (,$(findstring Linux,$(shell uname)))
    # Warning: make won't treat nested ifs as makefile directives if you use tabs instead of spaces
    ifneq (,$(findstring musl,$(shell ldd --version))) # linux (musl)
		CC=gcc CXX=g++ TAGS=musl $(MAKE) cli-builder
    else # linux (glibc)
		CC=gcc CXX=g++ $(MAKE) cli-builder
    endif
else # darwin
	$(MAKE) cli-builder
endif

# Cross-compile from darwin to any of the OS/Arch pairs below
.PHONY: cross-build
cross-build:
ifeq ($(GOARCH),arm64)
	ifeq ($(GOOS),linux) # linux/arm64
		CC=aarch64-linux-musl-gcc CXX=aarch64-linux-musl-g++ CGO_LDFLAGS="-static" TAGS=musl $(MAKE) cli-builder
	else # darwin/arm64
		$(MAKE) cli-builder
	endif
else
	ifeq ($(GOOS),windows) # windows/amd64
		CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ CGO_LDFLAGS="-fstack-protector -static" $(MAKE) cli-builder
	else ifeq ($(GOOS),linux) # linux/amd64
		CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ CGO_LDFLAGS="-static" TAGS=musl $(MAKE) cli-builder
	else # darwin/amd64
		$(MAKE) cli-builder
	endif
endif

.PHONY: cli-builder
cli-builder:
	GOOS="" GOARCH="" CC="" CXX="" CGO_LDFLAGS="" go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION)

ifeq ($(GOLANG_FIPS),1)
	wget "https://go.dev/dl/go$$(cat .go-version).src.tar.gz" && \
	tar -xf go$$(cat .go-version).src.tar.gz && \
	git clone --branch go$$(cat .go-version)-1-openssl-fips --depth 1 https://github.com/golang-fips/go.git go-openssl && \
	cd go/ && \
	cat ../go-openssl/patches/*.patch | patch -p1 && \
	sed -i '' 's/linux/darwin/' src/crypto/internal/backend/nobackend.go && \
	sed -i '' 's/linux/darwin/' src/crypto/internal/backend/openssl.go && \
	sed -i '' 's/"libcrypto.so.%s"/"libcrypto.%s.dylib"/' src/crypto/internal/backend/openssl.go && \
	cd src/ && \
	./make.bash && \
	cd ../../
	PATH=$$(pwd)/go/bin:$$PATH GOROOT=$$(pwd)/go TAGS=$(TAGS) CC=$(CC) CXX=$(CXX) CGO_LDFLAGS=$(CGO_LDFLAGS) goreleaser build --clean --single-target --snapshot
	rm -rf go go-openssl go$$(cat .go-version).src.tar.gz
else
	TAGS=$(TAGS) CC=$(CC) CXX=$(CXX) CGO_LDFLAGS=$(CGO_LDFLAGS) goreleaser build --clean --single-target --snapshot
endif



.PHONY: clean
clean:
	for dir in bin dist docs legal prebuilt release-notes; do \
		[ -d $$dir ] && rm -r $$dir || true; \
	done

.PHONY: lint
lint: lint-go lint-cli

.PHONY: lint-go
lint-go:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 && \
	golangci-lint run --timeout 10m
	@echo "✅  golangci-lint"

.PHONY: lint-cli
lint-cli: cmd/lint/en_US.aff cmd/lint/en_US.dic
	go run cmd/lint/main.go -aff-file $(word 1,$^) -dic-file $(word 2,$^) $(ARGS)
	@echo "✅  cmd/lint/main.go"

cmd/lint/en_US.aff:
	curl -s "https://chromium.googlesource.com/chromium/deps/hunspell_dictionaries/+/master/en_US.aff?format=TEXT" | base64 -D > $@

cmd/lint/en_US.dic:
	curl -s "https://chromium.googlesource.com/chromium/deps/hunspell_dictionaries/+/master/en_US.dic?format=TEXT" | base64 -D > $@

.PHONY: unit-test
unit-test:
	@if [ ! -f "$(shell go env GOPATH)/bin/gotestsum$(shell go env GOEXE)" ]; then \
		go install gotest.tools/gotestsum@v1.12.1 ; \
	fi
	@"$(shell go env GOPATH)/bin/gotestsum$(shell go env GOEXE)" --junitfile unit-test-report.xml -- \
		-timeout 0 -v -race -coverprofile=coverage.unit.out -covermode=atomic \
		$$(go list ./... | grep -v github.com/confluentinc/cli/v4/test)

.PHONY: build-for-integration-test
build-for-integration-test:
ifdef CI
	go build -cover -ldflags="-s -w -X main.commit="00000000" -X main.date="1970-01-01T00:00:00Z" -X main.isTest=true" -o test/bin/confluent ./cmd/confluent
else
	go build -cover -ldflags="-s -w -X main.commit="00000000" -X main.date="1970-01-01T00:00:00Z" -X main.isTest=true" -o test/bin/confluent ./cmd/confluent
endif

.PHONY: build-for-integration-test-windows
build-for-integration-test-windows:
ifdef CI
	go build -cover -ldflags="-s -w -X main.commit="00000000" -X main.date="1970-01-01T00:00:00Z" -X main.isTest=true" -o test/bin/confluent.exe ./cmd/confluent
else
	go build -cover -ldflags="-s -w -X main.commit="00000000" -X main.date="1970-01-01T00:00:00Z" -X main.isTest=true" -o test/bin/confluent.exe ./cmd/confluent
endif

.PHONY: integration-test
integration-test:
	@if [ ! -f "$(shell go env GOPATH)/bin/gotestsum$(shell go env GOEXE)" ]; then \
		go install gotest.tools/gotestsum@v1.12.1 ; \
	fi
	@rm -rf test/coverage && mkdir -p test/coverage && \
	GOCOVERDIR=test/coverage "$(shell go env GOPATH)/bin/gotestsum$(shell go env GOEXE)" --junitfile integration-test-report.xml -- \
		-timeout 0 -v -race $$(go list ./... | grep github.com/confluentinc/cli/v4/test) && \
	go tool covdata textfmt -i test/coverage -o coverage.integration.out


.PHONY: test
test: unit-test integration-test

.PHONY: generate-packaging-patch
generate-packaging-patch:
	diff -u Makefile debian/Makefile | sed "1 s_Makefile_cli/Makefile_" > debian/patches/standard_build_layout.patch

.PHONY: coverage
coverage:
	@echo "Merging coverage data..."
	@echo "mode: atomic" > coverage.out
	@tail -n +2 coverage.unit.out >> coverage.out || echo "No unit coverage found"
	@tail -n +2 coverage.integration.out >> coverage.out || echo "No integration coverage found"

ifeq ($(CI),true)
	@cp coverage.out coverage.txt
	@echo "Coverage data saved to: coverage.txt"
	@artifact push workflow coverage.txt
else
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
endif
