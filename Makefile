SHELL := /bin/bash
GORELEASER_VERSION := v2.13.3

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
	GOOS="" GOARCH="" CC="" CXX="" CGO_LDFLAGS="" go install github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION)

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
ifdef CI
	go install gotest.tools/gotestsum@v1.13.0 && \
	gotestsum --junitfile unit-test-report.xml -- -timeout 0 -v -race -coverprofile=coverage.unit.out -covermode=atomic $$(go list ./... | grep -v github.com/confluentinc/cli/v4/test)
else
	go test -timeout 0 -v -coverprofile=coverage.unit.out -covermode=atomic $$(go list ./... | grep -v github.com/confluentinc/cli/v4/test) $(UNIT_TEST_ARGS)
endif

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
ifdef CI
	go install gotest.tools/gotestsum@v1.13.0 && \
	export GOCOVERDIR=test/coverage && \
	rm -rf $${GOCOVERDIR} && mkdir $${GOCOVERDIR} && \
	gotestsum --junitfile integration-test-report.xml -- -timeout 0 -v -race $$(go list ./... | grep github.com/confluentinc/cli/v4/test) && \
	go tool covdata textfmt -i $${GOCOVERDIR} -o coverage.integration.out
else
	export GOCOVERDIR=test/coverage && \
	rm -rf $${GOCOVERDIR} && mkdir $${GOCOVERDIR} && \
	go test -timeout 0 -v $$(go list ./... | grep github.com/confluentinc/cli/v4/test) $(INTEGRATION_TEST_ARGS) && \
	go tool covdata textfmt -i $${GOCOVERDIR} -o coverage.integration.out
endif

.PHONY: test
test: unit-test integration-test

.PHONY: build-for-live-test
build-for-live-test:
	go build -ldflags="-s -w -X main.disableUpdates=true" -o test/live/bin/confluent ./cmd/confluent

.PHONY: live-test
live-test: build-for-live-test
	@if [ -z "$(CLI_LIVE_TEST_GROUPS)" ]; then \
		CLI_LIVE_TEST=1 go test ./test/live/ -v -run=".*Live$$" \
			-tags="live_test,all" -timeout 1440m -parallel 10; \
	else \
		TAGS="live_test"; \
		for group in $$(echo "$(CLI_LIVE_TEST_GROUPS)" | tr ',' ' '); do \
			TAGS="$$TAGS,$$group"; \
		done; \
		CLI_LIVE_TEST=1 go test ./test/live/ -v -run=".*Live$$" \
			-tags="$$TAGS" -timeout 1440m -parallel 10; \
	fi

.PHONY: live-test-core
live-test-core:
	@$(MAKE) live-test CLI_LIVE_TEST_GROUPS="core"

.PHONY: live-test-kafka
live-test-kafka:
	@$(MAKE) live-test CLI_LIVE_TEST_GROUPS="kafka"

.PHONY: live-test-schema-registry
live-test-schema-registry:
	@$(MAKE) live-test CLI_LIVE_TEST_GROUPS="schema_registry"

.PHONY: live-test-iam
live-test-iam:
	@$(MAKE) live-test CLI_LIVE_TEST_GROUPS="iam"

.PHONY: live-test-auth
live-test-auth:
	@$(MAKE) live-test CLI_LIVE_TEST_GROUPS="auth"

.PHONY: live-test-connect
live-test-connect:
	@$(MAKE) live-test CLI_LIVE_TEST_GROUPS="connect"

.PHONY: live-test-billing
live-test-billing:
	@$(MAKE) live-test CLI_LIVE_TEST_GROUPS="billing"

.PHONY: live-test-flink
live-test-flink:
	@$(MAKE) live-test CLI_LIVE_TEST_GROUPS="flink"

.PHONY: live-test-essential
live-test-essential:
	@$(MAKE) live-test CLI_LIVE_TEST_GROUPS="core,kafka,schema_registry,auth,billing"

.PHONY: live-test-multicloud
live-test-multicloud:
	@CLI_LIVE_TEST_VARIANTS="aws:us-east-1:basic,gcp:us-east1:basic,azure:eastus:basic" \
		$(MAKE) live-test CLI_LIVE_TEST_GROUPS="kafka"

.PHONY: live-test-resource
live-test-resource: build-for-live-test
	@if [ -z "$(RESOURCE)" ]; then \
		echo "Available resources:"; \
		echo ""; \
		printf "  %-25s %-20s %s\n" "RESOURCE" "GROUP" "TEST FUNCTION"; \
		printf "  %-25s %-20s %s\n" "--------" "-----" "-------------"; \
		printf "  %-25s %-20s %s\n" "environment" "core" "TestEnvironmentCRUDLive"; \
		printf "  %-25s %-20s %s\n" "service_account" "core" "TestServiceAccountCRUDLive"; \
		printf "  %-25s %-20s %s\n" "api_key" "core" "TestApiKeyCRUDLive"; \
		printf "  %-25s %-20s %s\n" "kafka_cluster" "kafka" "TestKafkaClusterCRUDLive"; \
		printf "  %-25s %-20s %s\n" "kafka_topic" "kafka" "TestKafkaTopicCRUDLive"; \
		printf "  %-25s %-20s %s\n" "kafka_acl" "kafka" "TestKafkaACLCRUDLive"; \
		printf "  %-25s %-20s %s\n" "kafka_consumer_group" "kafka" "TestKafkaConsumerGroupListLive"; \
		printf "  %-25s %-20s %s\n" "schema_registry" "schema_registry" "TestSchemaRegistrySchemaCRUDLive"; \
		printf "  %-25s %-20s %s\n" "iam_rbac" "iam" "TestRBACRoleBindingCRUDLive"; \
		printf "  %-25s %-20s %s\n" "login" "auth" "TestLoginLogoutLive"; \
		printf "  %-25s %-20s %s\n" "connect" "connect" "TestConnectClusterCRUDLive"; \
		printf "  %-25s %-20s %s\n" "organization" "core" "TestOrganizationLive"; \
		printf "  %-25s %-20s %s\n" "audit_log" "core" "TestAuditLogLive"; \
		printf "  %-25s %-20s %s\n" "service_quota" "core" "TestServiceQuotaLive"; \
		printf "  %-25s %-20s %s\n" "context" "core" "TestContextAndConfigurationLive"; \
		printf "  %-25s %-20s %s\n" "billing" "billing" "TestBillingLive"; \
		printf "  %-25s %-20s %s\n" "kafka_region" "kafka" "TestKafkaRegionListLive"; \
		printf "  %-25s %-20s %s\n" "kafka_quota" "kafka" "TestKafkaQuotaCRUDLive"; \
		printf "  %-25s %-20s %s\n" "kafka_produce_consume" "kafka" "TestKafkaProduceConsumeLive"; \
		printf "  %-25s %-20s %s\n" "iam_user" "iam" "TestIAMUserLive"; \
		printf "  %-25s %-20s %s\n" "iam_ip_group_filter" "iam" "TestIAMIpGroupFilterCRUDLive"; \
		printf "  %-25s %-20s %s\n" "iam_identity_provider" "iam" "TestIAMIdentityProviderCRUDLive"; \
		printf "  %-25s %-20s %s\n" "iam_certificate" "iam" "TestIAMCertificateAuthorityCRUDLive"; \
		printf "  %-25s %-20s %s\n" "connect_custom_plugin" "connect" "TestConnectCustomPluginListLive"; \
		printf "  %-25s %-20s %s\n" "schema_registry_ext" "schema_registry" "TestSchemaRegistryExtendedLive"; \
		printf "  %-25s %-20s %s\n" "flink_region" "flink" "TestFlinkRegionListLive"; \
		printf "  %-25s %-20s %s\n" "plugin" "core" "TestPluginListLive"; \
		printf "  %-25s %-20s %s\n" "kafka_share_group" "kafka" "TestKafkaShareGroupListLive"; \
		echo ""; \
		echo "Usage: make live-test-resource RESOURCE=<resource>"; \
	else \
		case "$(RESOURCE)" in \
			environment) GROUP=core; FUNC=TestEnvironmentCRUDLive;; \
			service_account) GROUP=core; FUNC=TestServiceAccountCRUDLive;; \
			api_key) GROUP=core; FUNC=TestApiKeyCRUDLive;; \
			kafka_cluster) GROUP=kafka; FUNC=TestKafkaClusterCRUDLive;; \
			kafka_topic) GROUP=kafka; FUNC=TestKafkaTopicCRUDLive;; \
			kafka_acl) GROUP=kafka; FUNC=TestKafkaACLCRUDLive;; \
			kafka_consumer_group) GROUP=kafka; FUNC=TestKafkaConsumerGroupListLive;; \
			schema_registry) GROUP=schema_registry; FUNC=TestSchemaRegistrySchemaCRUDLive;; \
			iam_rbac) GROUP=iam; FUNC=TestRBACRoleBindingCRUDLive;; \
			login) GROUP=auth; FUNC=TestLoginLogoutLive;; \
			connect) GROUP=connect; FUNC=TestConnectClusterCRUDLive;; \
			organization) GROUP=core; FUNC=TestOrganizationLive;; \
			audit_log) GROUP=core; FUNC=TestAuditLogLive;; \
			service_quota) GROUP=core; FUNC=TestServiceQuotaLive;; \
			context) GROUP=core; FUNC=TestContextAndConfigurationLive;; \
			billing) GROUP=billing; FUNC=TestBillingLive;; \
			kafka_region) GROUP=kafka; FUNC=TestKafkaRegionListLive;; \
			kafka_quota) GROUP=kafka; FUNC=TestKafkaQuotaCRUDLive;; \
			kafka_produce_consume) GROUP=kafka; FUNC=TestKafkaProduceConsumeLive;; \
			iam_user) GROUP=iam; FUNC=TestIAMUserLive;; \
			iam_ip_group_filter) GROUP=iam; FUNC=TestIAMIpGroupFilterCRUDLive;; \
			iam_identity_provider) GROUP=iam; FUNC=TestIAMIdentityProviderCRUDLive;; \
			iam_certificate) GROUP=iam; FUNC=TestIAMCertificateAuthorityCRUDLive;; \
			connect_custom_plugin) GROUP=connect; FUNC=TestConnectCustomPluginListLive;; \
			schema_registry_ext) GROUP=schema_registry; FUNC=TestSchemaRegistryExtendedLive;; \
			flink_region) GROUP=flink; FUNC=TestFlinkRegionListLive;; \
			plugin) GROUP=core; FUNC=TestPluginListLive;; \
			kafka_share_group) GROUP=kafka; FUNC=TestKafkaShareGroupListLive;; \
			*) echo "Unknown resource: $(RESOURCE)"; echo "Run 'make live-test-resource' to see available resources."; exit 1;; \
		esac; \
		echo "Running $$FUNC (group: $$GROUP)..."; \
		CLI_LIVE_TEST=1 go test ./test/live/ -v -run="TestLive/$$FUNC" \
			-tags="live_test,$$GROUP" -timeout 1440m -parallel 10; \
	fi

.PHONY: generate-packaging-patch
generate-packaging-patch:
	diff -u Makefile debian/Makefile | sed "1 s_Makefile_cli/Makefile_" > debian/patches/standard_build_layout.patch

.PHONY: coverage
coverage: ## Merge coverage data from unit and integration tests into coverage.txt
	@echo "Merging coverage data..."
	@echo "mode: atomic" > coverage.txt
	@tail -n +2 coverage.unit.out >> coverage.txt
	@tail -n +2 coverage.integration.out >> coverage.txt
	@echo "Coverage data saved to: coverage.txt"
	@artifact push workflow coverage.txt